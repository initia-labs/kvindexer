package pair

import (
	"context"
	"errors"
	"time"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/types/tx"

	"github.com/initia-labs/kvindexer/module/keeper"
)

//nolint:unused
var height int64

//nolint:unused
var timestamp time.Time

func preparer(k *keeper.Keeper, ctx context.Context) (err error) {
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
	if k.NftTransferKeeper == nil {
		return errors.New("nft transfer keeper is not set")
	}

	return nil
}

func initializer(k *keeper.Keeper, ctx context.Context) (err error) {
	return nil
}

func parseTx(k *keeper.Keeper, txBytes []byte) (*tx.Tx, error) {
	tx := tx.Tx{}
	err := k.GetCodec().Unmarshal(txBytes, &tx)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func finalizeBlock(k *keeper.Keeper, ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	k.Logger(ctx).Debug("finalizeBlock", "submodule", submoduleName, "txs", len(req.Txs), "height", req.Height)

	// set these everytime: it'll be used in commit()
	// is okay to set height here because finalizeBlock is called before commit
	height = req.Height
	timestamp = req.Time

	if err := collectOPfungibleTokens(k, ctx, req); err != nil {
		k.Logger(ctx).Warn("collectOPfungibleTokens", "error", err, "submodule", submoduleName)
	}

	if err := collectIBCFungibleTokens(k, ctx); err != nil {
		// don't return error
		k.Logger(ctx).Warn("collectIBCFungibleTokens", "error", err, "submodule", submoduleName)
	}

	if err := collectIBCNonfungibleTokens(k, ctx, res); err != nil {
		k.Logger(ctx).Warn("collectIBCNonfungibleTokens", "error", err, "submodule", submoduleName)
	}

	return nil
}

func commit(k *keeper.Keeper, ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) error {
	k.Logger(ctx).Debug("commit", "submodule", submoduleName)

	// nop

	return nil
}
