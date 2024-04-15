package nft

import (
	"context"
	"time"

	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/initia-labs/kvindexer/module/keeper"
)

//nolint:unused
var height int64

//nolint:unused
var timestamp time.Time

func preparer(k *keeper.Keeper, ctx context.Context) (err error) {
	return addStorages(k, ctx)

}

func initializer(k *keeper.Keeper, ctx context.Context) (err error) {
	return nil
}

func finalizeBlock(k *keeper.Keeper, ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	k.Logger(ctx).Debug("finalizeBlock", "submodule", submoduleName, "txs", len(req.Txs), "height", req.Height)

	// set these everytime: it'll be used in commit()
	// is okay to set height here because finalizeBlock is called before commit
	height = req.Height
	timestamp = req.Time

	for _, txResult := range res.TxResults {
		events := filterAndParseEvent(txResult.Events, eventTypes)
		err := processEvents(k, ctx, events)
		if err != nil {
			k.Logger(ctx).Warn("failed to process events", "error", err, "submodule", submoduleName)
		}
		for _, event := range txResult.Events {
			if event.Type == "write_acknowledgement" {
				err := handleWriteAcknowledgementEvent(k, ctx, event.Attributes)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
func commit(k *keeper.Keeper, ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) error {
	k.Logger(ctx).Debug("commit", "submodule", submoduleName)
	// nop here
	return nil
}
