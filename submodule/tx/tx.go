package tx

import (
	"context"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/initia-labs/kvindexer/module/keeper"
)

//nolint:unused
var height int64

//nolint:unused
var timestamp time.Time

var fbreq abci.RequestFinalizeBlock
var fbres abci.ResponseFinalizeBlock

func preparer(k *keeper.Keeper, ctx context.Context) (err error) {
	cdc := k.GetCodec()

	if txMap, err = keeper.AddMap(k, prefixTxs, txsName, collections.StringKey, codec.CollValue[sdk.TxResponse](cdc)); err != nil {
		return err
	}

	if txhashesByAccountMap, err = keeper.AddMap(k, prefixTxsByAccount, txsPerByAccountName, collections.PairKeyCodec(sdk.AccAddressKey, collections.Uint64Key), collections.StringValue); err != nil {
		return err
	}
	if accountSequenceMap, err = keeper.AddMap(k, prefixAccountSequence, accountSequenceName, sdk.AccAddressKey, collections.Uint64Value); err != nil {
		return err
	}

	if txhashesBySequence, err = keeper.AddMap(k, prefixTxSequence, txSequenceName, collections.Uint64Key, collections.StringValue); err != nil {
		return err
	}

	if txhashesByHeight, err = keeper.AddMap(k, prefixTxByHeight, txByHeightName, collections.PairKeyCodec(collections.Int64Key, collections.Uint64Key), collections.StringValue); err != nil {
		return err
	}

	if sequence, err = keeper.AddSequence(k, prefixSequence, sequenceName); err != nil {
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

	if err := processTxs(k, ctx, fbreq, fbres); err != nil {
		return err
	}

	return nil
}
