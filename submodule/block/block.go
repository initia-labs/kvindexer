package block

import (
	"context"
	"time"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/initia-labs/kvindexer/config"
	"github.com/initia-labs/kvindexer/module/keeper"
	"github.com/spf13/cast"
)

var limit uint
var enabled bool

//nolint:unused
var height int64

//nolint:unused
var timestamp time.Time

func preparer(k *keeper.Keeper, ctx context.Context, cfg config.SubmoduleConfig) (err error) {
	enabled = true //assume that it passes handler's prepare func.

	if blockByHeight, err = keeper.AddMap(k, prefixBlock, blockByHeightName, collections.Uint64Key, collections.BytesValue); err != nil {
		return err
	}

	return nil
}

func initializer(k *keeper.Keeper, ctx context.Context, cfg config.SubmoduleConfig) (err error) {
	limit = cast.ToUint(cfg["limit"])
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

	if err := collectBlock(k, ctx, req, res, cfg); err != nil {
		return err
	}

	return nil
}
func commit(k *keeper.Keeper, ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair, cfg config.SubmoduleConfig) error {
	k.Logger(ctx).Debug("commit", "submodule", submoduleName)

	return nil
}
