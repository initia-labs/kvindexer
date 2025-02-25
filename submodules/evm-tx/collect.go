package tx

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/collections"
	cosmoserr "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"

	evmtypes "github.com/initia-labs/minievm/x/evm/types"

	"github.com/initia-labs/kvindexer/submodules/evm-tx/types"
)

const (
	transferTopic = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
)

func (sm EvmTxSubmodule) finalizeBlock(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	sm.Logger(ctx).Debug("finalizeBlock", "submodule", types.SubmoduleName, "txs", len(req.Txs), "height", req.Height)

	if err := sm.processTxs(ctx, req, res); err != nil {
		return err
	}

	return nil
}

func (sm EvmTxSubmodule) processTxs(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	// key: address, value: txs slice
	accTxMap := map[string][]string{}

	txHashes := []string{}
	for idx, txBytes := range req.Txs {
		tx, err := parseTx(sm.cdc, txBytes)
		if err != nil {
			sm.Logger(ctx).Info("failed to parse tx", "error", err, "index", idx)
			continue
		}

		any, err := codectypes.NewAnyWithValue(tx)
		if err != nil {
			sm.Logger(ctx).Info("failed to unpack any", "error", err, "index", idx)
			continue
		}

		txHash := tmhash.Sum(txBytes)
		txHashStr := fmt.Sprintf("%x", txHash)
		resultTx := coretypes.ResultTx{
			Hash:     txHash,
			Height:   req.Height,
			TxResult: *res.TxResults[idx],
			// No Index, Tx and Proof here. sdk.NewResponseTxResult() don't use them
		}

		txr := sdk.NewResponseResultTx(&resultTx, any, req.Time.UTC().Format(time.RFC3339))

		if err := sm.txMap.Set(ctx, txHashStr, *txr); err != nil {
			sm.Logger(ctx).Info("failed to store tx", "error", err, "index", idx)
			continue
		}

		// get addresses from the tx
		addrs, err := grepAddressesFromTx(txr)
		if err != nil {
			sm.Logger(ctx).Info("failed to grep addresses from tx", "error", err, "index", idx)
			continue
		}

		for _, addr := range addrs {
			accTxMap[addr] = uniqueAppend(accTxMap[addr], txHashStr)
		}
		txHashes = append(txHashes, txHashStr)
	}

	// store tx/account pair into txAccMap
	for addr, txHashes := range accTxMap {
		err := sm.storeAccTxs(ctx, req.Height, addr, txHashes)
		if err != nil {
			sm.Logger(ctx).Info("failed to store tx/account pair", "error", err, "address", addr)
		}
	}

	return sm.storeIndices(ctx, req.Height, txHashes)
}

func uniqueAppend(slice []string, elem string) []string {
	for _, e := range slice {
		if e == elem {
			return slice
		}
	}
	return append(slice, elem)
}

func parseTx(cdc codec.Codec, txBytes []byte) (*tx.Tx, error) {
	tx := tx.Tx{}
	err := cdc.Unmarshal(txBytes, &tx)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func grepAddressesFromTx(txr *sdk.TxResponse) ([]string, error) {
	grepped := []string{}
	for _, event := range txr.Events {
		for _, attr := range event.Attributes {
			var addrs []string

			switch {
			case event.Type == evmtypes.EventTypeEVM && attr.Key == evmtypes.AttributeKeyLog:
				contractAddrs, err := extractAddressesFromEVMLog(attr.Value)
				if err != nil {
					continue
				}
				addrs = append(addrs, contractAddrs...)
			case isEvmModuleEvent(event.Type) && attr.Key == evmtypes.AttributeKeyContract:
				addr, err := convertContractAddressToBech32(attr.Value)
				if err != nil {
					continue
				}
				addrs = append(addrs, addr)
			default:
				addrs = findAllBech32Address(attr.Value)
				addrs = append(addrs, findAllHexAddress(attr.Value)...)
			}

			for _, addr := range addrs {
				accAddr, err := accAddressFromString(addr)
				if err != nil {
					continue
				}
				grepped = append(grepped, accAddr.String())
			}
		}
	}

	return grepped, nil
}

// isEvmModuleEvent checks if the event type is from evm module except evmtypes.EventTypeEVM.
// return true if it is, false otherwise.
func isEvmModuleEvent(eventType string) bool {
	switch eventType {
	case evmtypes.EventTypeCall, evmtypes.EventTypeCreate,
		evmtypes.EventTypeContractCreated, evmtypes.EventTypeERC20Created,
		evmtypes.EventTypeERC721Created, evmtypes.EventTypeERC721Minted, evmtypes.EventTypeERC721Burned:
		return true
	default:
		return false
	}
}

func extractAddressesFromEVMLog(attrVal string) (addrs []string, err error) {
	log := evmtypes.Log{}
	if err = json.Unmarshal([]byte(attrVal), &log); err != nil {
		return
	}
	var addr string
	addr, err = convertContractAddressToBech32(log.Address)
	if err == nil {
		addrs = append(addrs, addr)
	}

	// if the topic is about transfer, we need to extract the addresses from the topics.
	if log.Topics == nil { // no topic
		return
	}
	topicLen := len(log.Topics)
	if topicLen < 2 { // no data to extract
		return
	}
	if log.Topics[0] != transferTopic { // topic is not about transfer
		return
	}

	for i := 1; i < topicLen; i++ {
		if i == 3 { // if index is 3, it means index indicates the amount, not address. need break
			break
		}
		addr, err = convertContractAddressToBech32(log.Topics[i])
		if err != nil {
			continue
		}
		addrs = append(addrs, addr)
	}

	return
}

func (sm EvmTxSubmodule) storeAccTxs(ctx context.Context, height int64, addr string, txHashes []string) error {
	if len(txHashes) == 0 {
		return nil
	}
	acc, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		sm.Logger(ctx).Info("failed to convert address", "error", err, "address", addr)
		return err
	}

	seq, err := sm.accountSequenceMap.Get(ctx, acc)
	if err != nil && !cosmoserr.IsOf(err, collections.ErrNotFound) {
		return err
	}

	for i, txHash := range txHashes {
		err = sm.txhashesByAccountMap.Set(ctx, collections.Join(acc, seq+uint64(i)), txHash)
		if err != nil {
			sm.Logger(ctx).Info("failed to store tx/account pair", "error", err, "address", addr, "txhash", txHash)
			continue
		}
	}

	delta := seq + uint64(len(txHashes)+1)
	if err = sm.accountSequenceMap.Set(ctx, acc, delta); err != nil {
		sm.Logger(ctx).Info("failed to store account sequence", "error", err, "address", addr, "delta", delta)
		return err
	}

	// store (height, account, sequence) for pruning
	return sm.accountSequenceByHeightMap.Set(ctx, collections.Join3(height, acc, delta), true)
}

func (sm EvmTxSubmodule) storeIndices(ctx context.Context, height int64, txHashes []string) error {
	for i, txHash := range txHashes {
		err := sm.txhashesByHeightMap.Set(ctx, collections.Join(height, uint64(i)), txHash)
		if err != nil {
			sm.Logger(ctx).Info("failed to store tx/height pair", "error", err, "height", height, "txhash", txHash)
			continue
		}

		seq, err := sm.sequence.Next(ctx)
		if err != nil {
			return err
		}

		err = sm.txhashesBySequenceMap.Set(ctx, seq, txHash)
		if err != nil {
			return err
		}
	}

	// store height -> sequence for pruning
	seq, err := sm.sequence.Peek(ctx)
	if err != nil {
		return err
	}

	return sm.sequenceByHeightMap.Set(ctx, height, seq)
}
