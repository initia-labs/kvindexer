package nft

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/initia-labs/kvindexer/module/keeper"
)

func preparer(k *keeper.Keeper, ctx context.Context) (err error) {
	return addStorages(k, ctx)

}

func initializer(k *keeper.Keeper, ctx context.Context) (err error) {
	return nil
}

func finalizeBlock(k *keeper.Keeper, ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	k.Logger(ctx).Debug("finalizeBlock", "submodule", submoduleName, "txs", len(req.Txs), "height", req.Height)

	for _, txResult := range res.TxResults {
		events := filterAndParseEvent(eventType, txResult.Events)
		err := processEvents(k, ctx, events)
		if err != nil {
			k.Logger(ctx).Warn("processEvents", "error", err)
		}
	}

	return nil
}

func commit(k *keeper.Keeper, ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) error {
	k.Logger(ctx).Debug("commit", "submodule", submoduleName)

	return nil
}
