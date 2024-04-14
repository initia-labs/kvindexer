package nft

import (
	"context"
	"time"

	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/spf13/cast"

	"github.com/initia-labs/kvindexer/config"
	"github.com/initia-labs/kvindexer/module/keeper"
)

var limit uint64
var enabled bool

//nolint:unused
var height int64

//nolint:unused
var timestamp time.Time

func preparer(k *keeper.Keeper, ctx context.Context, cfg config.SubmoduleConfig) (err error) {
	enabled = true //assume that it passes handler's prepare func.

	return addStorages(k, ctx, cfg)

}

func initializer(k *keeper.Keeper, ctx context.Context, cfg config.SubmoduleConfig) (err error) {
	limit = cast.ToUint64(cfg["limit"])
	if limit == 0 {
		limit = 100
	}
	return nil
}

func finalizeBlock(k *keeper.Keeper, ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock, cfg config.SubmoduleConfig) error {
	k.Logger(ctx).Debug("finalizeBlock", "submodule", submoduleName, "txs", len(req.Txs), "height", req.Height)

	// set these everytime: it'll be used in commit()
	// is okay to set height here because finalizeBlock is called before commit
	height = req.Height
	timestamp = req.Time

	for _, txResult := range res.TxResults {
		events := filterAndParseEvent(txResult.Events, eventTypes)
		err := processEvents(k, ctx, cfg, events)
		if err != nil {
			return err
		}
		for _, event := range txResult.Events {
			if event.Type == "write_acknowledgement" {
				err := handleWriteAcknowledgementEvent(k, ctx, cfg, event.Attributes)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
func commit(k *keeper.Keeper, ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair, cfg config.SubmoduleConfig) error {
	k.Logger(ctx).Debug("commit", "submodule", submoduleName)
	// nop here
	return nil
}
