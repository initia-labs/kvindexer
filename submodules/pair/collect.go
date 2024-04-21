package pair

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

	exportedibc "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"

	opchildtypes "github.com/initia-labs/OPinit/x/opchild/types"
	"github.com/initia-labs/kvindexer/submodules/pair/types"
)

const (
	ibcTransferPort = "transfer"
)

func (sm PairSubmodule) finalizeBlock(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	sm.Logger(ctx).Debug("finalizeBlock", "submodule", types.SubmoduleName, "txs", len(req.Txs), "height", req.Height)

	if err := sm.collectOPfungibleTokens(ctx, req); err != nil {
		sm.Logger(ctx).Warn("collectOPfungibleTokens", "error", err, "submodule", types.SubmoduleName)
	}

	if err := sm.collectIBCFungibleTokens(ctx); err != nil {
		// don't return error
		sm.Logger(ctx).Warn("collectIBCFungibleTokens", "error", err, "submodule", types.SubmoduleName)
	}

	if err := sm.collectIBCNonfungibleTokens(ctx, res); err != nil {
		sm.Logger(ctx).Warn("collectIBCNonfungibleTokens", "error", err, "submodule", types.SubmoduleName)
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

func (sm PairSubmodule) collectOPfungibleTokens(ctx context.Context, req abci.RequestFinalizeBlock) (err error) {
	for _, txBytes := range req.Txs {
		tx, err := sm.parseTx(txBytes)
		if err != nil {
			return err
		}
		for _, msg := range tx.GetMsgs() {
			targetMsg, ok := msg.(*opchildtypes.MsgFinalizeTokenDeposit)
			if !ok {
				continue
			}
			err := sm.SetPair(ctx, false, true, targetMsg.Amount.Denom, targetMsg.BaseDenom)
			if err != nil {
				sm.Logger(ctx).Warn("SetPair", "error", err, "denom", targetMsg.Amount.Denom, "baseDenom", targetMsg.BaseDenom)
			}
		}
	}
	return nil
}

func (sm PairSubmodule) collectIBCNonfungibleTokens(ctx context.Context, res abci.ResponseFinalizeBlock) (err error) {
	for _, txResult := range res.TxResults {
		for _, event := range txResult.Events {
			if event.Type != "write_acknowledgement" {
				continue
			}
			err := sm.handleWriteAcknowledgementEvent(ctx, event.Attributes)
			if err != nil {
				sm.Logger(ctx).Warn("failed to handle write_acknowledgement event", "error", err, "event", event)
			}
		}
	}
	return nil
}

func (sm PairSubmodule) handleWriteAcknowledgementEvent(ctx context.Context, attrs []abci.EventAttribute) (err error) {
	sm.Logger(ctx).Debug("write-ack", "attrs", attrs)
	for _, attr := range attrs {
		if attr.Key != "packet_data" {
			continue
		}

		data := types.WriteAckForNftEvent{}
		if err = json.Unmarshal([]byte(attr.Value), &data); err != nil {
			// may be not target
			return nil
		}

		cdb, err := base64.StdEncoding.DecodeString(data.ClassData)
		if err != nil {
			return cosmoserr.Wrap(err, "failed to decode class data")
		}
		classData := types.NftClassData{}
		if err = json.Unmarshal(cdb, &classData); err != nil {
			return cosmoserr.Wrap(err, "failed to unmarshal class data")
		}

		_, err = sm.GetPair(ctx, false, data.ClassId)
		if err == nil {
			return nil // already exists
		}
		if !cosmoserr.IsOf(err, collections.ErrNotFound) {
			return cosmoserr.Wrap(err, "failed to check class existence")
		}

		err = sm.SetPair(ctx, false, false, data.ClassId, classData.Description.Value)
		if err != nil {
			return cosmoserr.Wrap(err, "failed to set class")
		}

		sm.Logger(ctx).Info("nft class added", "classId", data.ClassId, "description", classData.Description.Value)
	}
	return nil
}

func (sm PairSubmodule) collectIBCFungibleTokens(ctx context.Context) error {

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	chKeeper := sm.channelKeeper

	// get all channels
	channels := chKeeper.GetAllChannels(sdkCtx)

	ibcChannels := []string{}
	for _, channel := range channels {
		if channel.PortId != ibcTransferPort {
			continue
		}

		_ /*clientId*/, cs, err := chKeeper.GetChannelClientState(sdkCtx, channel.PortId, channel.ChannelId)
		if err != nil {
			sm.Logger(ctx).Warn("GetChannelClientState", "error", err)
		}
		counterpartyChainId := getChainIdFromClientState(cs)
		if counterpartyChainId == "" {
			sm.Logger(ctx).Warn("channel id is nil")
			continue
		}
		if counterpartyChainId != sdk.UnwrapSDKContext(ctx).ChainID() {
			continue
		}
		ibcChannels = append(ibcChannels, channel.ChannelId)
	}

	traces := sm.transferKeeper.GetAllDenomTraces(sdkCtx)
	for _, ibcChannel := range ibcChannels {
		for _, trace := range traces {
			if trace.Path != fmt.Sprintf("%s/%s", ibcTransferPort, ibcChannel) {
				continue
			}

			prevDenom, err := sm.fungiblePairsMap.Get(ctx, trace.IBCDenom())
			if err != nil && !cosmoserr.IsOf(err, collections.ErrNotFound) {
				continue
			}
			// prevDenom should be empty string if not found, or already set
			if prevDenom == trace.BaseDenom {
				continue
			}

			err = sm.fungiblePairsMap.Set(ctx, trace.IBCDenom(), trace.BaseDenom)
			if err != nil {
				return err
			}
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
