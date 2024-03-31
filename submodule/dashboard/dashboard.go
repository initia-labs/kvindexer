package dashboard

import (
	"context"
	"time"

	"cosmossdk.io/collections"

	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/spf13/cast"

	"github.com/initia-labs/kvindexer/config"
	"github.com/initia-labs/kvindexer/module/keeper"
)

var limit uint64
var enabled bool

var height int64
var timestamp time.Time

func preparer(k *keeper.Keeper, ctx context.Context, cfg config.SubmoduleConfig) (err error) {
	enabled = true //assume that it passes handler's prepare func.

	if accountMapByHeight, err = keeper.AddMap(k, prefixAccountMapByHeight, newAccsByHeightName, collections.Int64Key, collections.StringValue); err != nil {
		return err
	}
	if newAccountCountMapByDate, err = keeper.AddMap(k, prefixAccountCountMapByDate, newAccsCountByDateName, collections.StringKey, collections.Uint64Value); err != nil {
		return err
	}
	if totalAccountCountByDate, err = keeper.AddMap(k, prefixTotalAccountCountByDate, totalAccsCountByDateName, collections.StringKey, collections.Uint64Value); err != nil {
		return err
	}
	if txCountByDate, err = keeper.AddMap(k, prefixTxCountByDate, txCountByDateKeyName, collections.StringKey, collections.Uint64Value); err != nil {
		return err
	}
	if supplyByDate, err = keeper.AddMap(k, prefixSupplyByDate, supplyByDateKeyName, collections.StringKey, collections.Uint64Value); err != nil {
		return err
	}
	if txTotalCountByDate, err = keeper.AddMap(k, prefixTxTotalCountByDate, txTotalCountByDateKeyName, collections.StringKey, collections.Uint64Value); err != nil {
		return err
	}

	return nil
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

	if err := processTxs(k, ctx, req, res, cfg); err != nil {
		return err
	}
	if err := processSupply(k, ctx, req, res, cfg); err != nil {
		return err
	}

	return nil
}
func commit(k *keeper.Keeper, ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair, cfg config.SubmoduleConfig) error {
	k.Logger(ctx).Debug("commit", "submodule", submoduleName)
	if err := processAccounts(k, ctx, res, changeSet, cfg); err != nil {
		return err
	}

	return nil
}
