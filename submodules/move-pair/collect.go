package move_pair

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"cosmossdk.io/collections"
	cosmoserr "cosmossdk.io/errors"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"

	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	exportedibc "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"

	opchildtypes "github.com/initia-labs/OPinit/x/opchild/types"
	"github.com/initia-labs/kvindexer/submodules/move_pair/types"
)

const (
	ibcTransferPort    = "transfer"
	ibcNftTransferPort = "nft-transfer"
)

func (sm PairSubmodule) finalizeBlock(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	sm.Logger(ctx).Debug("finalizeBlock", "submodule", types.SubmoduleName, "txs", len(req.Txs), "height", req.Height)

	if err := sm.collectIBCFungibleTokens(ctx); err != nil {
		// don't return error
		sm.Logger(ctx).Warn("collectIBCFungibleTokens", "error", err, "submodule", types.SubmoduleName)
	}

	for txIdx, txBytes := range req.Txs {
		tx, err := sm.parseTx(txBytes)
		if err != nil {
			return err
		}
		for _, msg := range tx.GetMsgs() {
			switch msg := msg.(type) {
			case *opchildtypes.MsgFinalizeTokenDeposit:
				err = sm.collectOPfungibleTokens(ctx, msg)
				if err != nil {
					sm.Logger(ctx).Warn("collectOPfungibleTokens", "error", err)
				}
			case *channeltypes.MsgRecvPacket:
				err = sm.collectIBCNonfungibleTokens(ctx, res.TxResults[txIdx])
				if err != nil {
					sm.Logger(ctx).Warn("collectIBCNonfungibleTokens", "error", err)
				}
			}
		}
	}

	return nil
}

func (sm PairSubmodule) parseTx(txBytes []byte) (*tx.Tx, error) {
	tx := tx.Tx{}
	err := sm.cdc.Unmarshal(txBytes, &tx)
	if err != nil {
		return nil, err
	}

	return &tx, nil
}

func (sm PairSubmodule) collectOPfungibleTokens(ctx context.Context, msg *opchildtypes.MsgFinalizeTokenDeposit) (err error) {
	err = sm.SetPair(ctx, false, true, msg.Amount.Denom, msg.BaseDenom)
	if err != nil {
		sm.Logger(ctx).Warn("SetPair", "error", err, "denom", msg.Amount.Denom, "baseDenom", msg.BaseDenom)
	}
	return nil
}

func (sm PairSubmodule) collectIBCNonfungibleTokens(ctx context.Context, txResult *abci.ExecTxResult) (err error) {
	var packetData, classId string

	for _, event := range txResult.Events {
		if event.Type == "recv_packet" {
			if sm.pickAttribute(event.Attributes, "packet_src_port") != ibcNftTransferPort {
				continue
			}
			packetData = sm.pickAttribute(event.Attributes, "packet_data")
			packetDstPort := sm.pickAttribute(event.Attributes, "packet_dst_port")
			packetDstChannel := sm.pickAttribute(event.Attributes, "packet_dst_channel")
			err = sm.pricessIbcNftPairEvent(ctx, packetData, packetDstPort, packetDstChannel)
			if err != nil {
				sm.Logger(ctx).Warn("failed to handle recv_packet event", "error", err, "recv_packet.packet_data", packetData, "class_trace.class_id", classId)
			}
		}
	}

	return nil
}

func (sm PairSubmodule) pickAttribute(attrs []abci.EventAttribute, key string) string {
	for _, attr := range attrs {
		if attr.Key == key {
			return attr.Value
		}
	}
	return ""
}

func (sm PairSubmodule) generateCw721FromIcs721PortInfo(port, channel string) string {
	return port + "/" + channel
}

func (sm PairSubmodule) pricessIbcNftPairEvent(ctx context.Context, packetDataStr, packetDstPort, packetDstChannel string) (err error) {
	sm.Logger(ctx).Debug("processPairEvent", "packet_data", packetDataStr, "port", packetDstPort, "channel", packetDstChannel)

	packetData := types.PacketData{}
	if err = json.Unmarshal([]byte(packetDataStr), &packetData); err != nil {
		// may be not target
		return nil
	}

	cdb, err := base64.StdEncoding.DecodeString(packetData.ClassData)
	if err != nil {
		return cosmoserr.Wrap(err, "failed to decode class data")
	}
	classData := types.NftClassData{}
	if err = json.Unmarshal(cdb, &classData); err != nil {
		return cosmoserr.Wrap(err, "failed to unmarshal class data")
	}

	collectionName := fmt.Sprintf("%s/%s", sm.generateCw721FromIcs721PortInfo(packetDstPort, packetDstChannel), packetData.ClassId)

	_, err = sm.GetPair(ctx, false, collectionName)
	if err == nil {
		return nil // already exists
	}
	if !cosmoserr.IsOf(err, collections.ErrNotFound) {
		return cosmoserr.Wrap(err, "failed to check class existence")
	}

	err = sm.SetPair(ctx, false, false, collectionName, classData.Name)
	if err != nil {
		return cosmoserr.Wrap(err, "failed to set class")
	}

	sm.Logger(ctx).Info("nft class added", "classId", collectionName, "classData", classData)
	return nil
}

func (sm PairSubmodule) collectIBCFungibleTokens(ctx context.Context) error {

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	traces := sm.transferKeeper.GetAllDenomTraces(sdkCtx)
	for _, trace := range traces {
		_, err := sm.fungiblePairsMap.Get(ctx, trace.IBCDenom())
		if err != nil && !cosmoserr.IsOf(err, collections.ErrNotFound) {
			continue
		}
		err = sm.fungiblePairsMap.Set(ctx, trace.IBCDenom(), trace.BaseDenom)
		if err != nil {
			return err
		}
	}

	return nil
}

func getChainIdFromClientState(csi exportedibc.ClientState) string {
	if csi == nil {
		return ""
	}
	cs, ok := csi.(*ibctm.ClientState)
	if !ok {
		return ""
	}
	return cs.ChainId
}
