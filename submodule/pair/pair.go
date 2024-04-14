package pair

import (
	"context"
	"errors"
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

//nolint:unused
var height int64

//nolint:unused
var timestamp time.Time

func preparer(k *keeper.Keeper, ctx context.Context, cfg config.SubmoduleConfig) (err error) {
	enabled = true //assume that it passes handler's prepare func.

	if nonFungiblePairsMap, err = keeper.AddMap(k, prefixNonFungiblePairs, nonFungiblePairMapName, collections.StringKey, collections.StringValue); err != nil {
		return err
	}
	if fungiblePairsMap, err = keeper.AddMap(k, prefixFungiblePairs, fungiblePairMapName, collections.StringKey, collections.StringValue); err != nil {
		return err
	}

	if k.IBCKeeper == nil {
		return errors.New("ibc keeper is not set")
	}
	if k.TransferKeeper == nil {
		return errors.New("transfer keeper is not set")
	}

	//cfg, err = getCronConfigFromSubmoduleConfig(config.SubmoduleConfig(cfg))
	croncfg, err = getCronConfigFromSubmoduleConfig(cfg)
	if err != nil {
		return err
	}

	err = k.RegisterCronjobWithPattern(croncfg.l1QueryPattern, Cronjob)
	if err != nil {
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

	if err := updateIBCChannels(k, ctx); err != nil {
		// don't return error
		k.Logger(ctx).Info("updateIBCChannels", "error", err)
	}

	if err := collectIbcTokenPairs(k, ctx); err != nil {
		return err
	}

	if err := collecOpTokenPairs(k, ctx); err != nil {
		return err
	}

	return nil
}

func commit(k *keeper.Keeper, ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair, cfg config.SubmoduleConfig) error {
	k.Logger(ctx).Debug("commit", "submodule", submoduleName)

	// nop

	return nil
}
