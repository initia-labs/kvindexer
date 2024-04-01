package tx

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/collections"
	cosmoserr "cosmossdk.io/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tx "github.com/cosmos/cosmos-sdk/types/tx"

	"github.com/initia-labs/kvindexer/config"
	"github.com/initia-labs/kvindexer/module/keeper"
)

func processTxs(k *keeper.Keeper, ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock, cfg config.SubmoduleConfig) error {
	// key: address, value: txs slice
	accTxMap := map[string][]string{}

	txHashes := []string{}
	for idx, txBytes := range req.Txs {
		tx, err := parseTx(k, txBytes)
		if err != nil {
			return err
		}

		any, err := codectypes.NewAnyWithValue(tx)
		if err != nil {
			return err
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

		if err := txMap.Set(ctx, txHashStr, *txr); err != nil {
			return err
		}

		// get addresses from the tx
		addrs, err := grepAddressesFromTx(txr)
		if err != nil {
			return err
		}

		for _, addr := range addrs {
			accTxMap[addr] = uniqueAppend(accTxMap[addr], txHashStr)
		}
		txHashes = append(txHashes, txHashStr)
	}

	// store tx/account pair into txAccMap
	for addr, txHashes := range accTxMap {
		err := storeAccTxs(ctx, addr, txHashes)
		if err != nil {
			return err
		}
	}
	return storeIndices(ctx, req.Height, txHashes)
}

func uniqueAppend(slice []string, elem string) []string {
	for _, e := range slice {
		if e == elem {
			return slice
		}
	}
	return append(slice, elem)
}

func parseTx(k *keeper.Keeper, txBytes []byte) (*tx.Tx, error) {
	tx := tx.Tx{}
	err := k.GetCodec().Unmarshal(txBytes, &tx)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func grepAddressesFromTx(txr *sdk.TxResponse) ([]string, error) {
	grepped := []string{}
	for _, event := range txr.Events {
		for _, attr := range event.Attributes {
			addrs := findAllBech32Address(attr.Value)
			addrs = append(addrs, findAllHexAddress(attr.Value)...)
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

func storeAccTxs(ctx context.Context, addr string, txHashes []string) error {
	if len(txHashes) == 0 {
		return nil
	}
	acc, _ := sdk.AccAddressFromBech32(addr)

	seq, err := accountSequenceMap.Get(ctx, acc)
	if err != nil && !cosmoserr.IsOf(err, collections.ErrNotFound) {
		return err
	}

	for i, txHash := range txHashes {
		err = txhashesByAccountMap.Set(ctx, collections.Join(acc, seq+uint64(i)), txHash)
		if err != nil {
			return err
		}

	}
	if err := accountSequenceMap.Set(ctx, acc, seq+uint64(len(txHashes)+1)); err != nil {
		return err
	}

	return nil
}

func storeIndices(ctx context.Context, height int64, txHashes []string) error {

	for i, txHash := range txHashes {
		err := txhashesByHeight.Set(ctx, collections.Join(height, uint64(i)), txHash)
		if err != nil {
			return err
		}

		seq, err := sequence.Next(ctx)
		if err != nil {
			return err
		}
		err = txhashesBySequence.Set(ctx, seq, txHash)
		if err != nil {
			return err
		}
	}

	return nil
}
