package block

import (
	"context"
	"time"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/initia-labs/kvindexer/module/keeper"
	"github.com/initia-labs/kvindexer/submodule/block/types"
)

//nolint:unused
var height int64

//nolint:unused
var timestamp time.Time

var fbreq abci.RequestFinalizeBlock
var fbres abci.ResponseFinalizeBlock

func preparer(k *keeper.Keeper, ctx context.Context) (err error) {
	cdc := k.GetCodec()

	if blockByHeight, err = keeper.AddMap(k, prefixBlock, blockByHeightName, collections.Int64Key, codec.CollValue[types.Block](cdc)); err != nil {
		return err
	}

	return nil
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
	fbreq = req
	fbres = res

	return nil
}
func commit(k *keeper.Keeper, ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) error {
	k.Logger(ctx).Debug("commit", "submodule", submoduleName)

	if err := collectBlock(k, ctx, fbreq, fbres); err != nil {
		return err
	}

	return nil
}
