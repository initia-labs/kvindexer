package nop

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/initia-labs/kvindexer/module/keeper"
)

func preparer(k *keeper.Keeper, ctx context.Context) error {
	k.Logger(ctx).Info("nop preparer")
	return nil
}

func initializer(k *keeper.Keeper, ctx context.Context) error {
	k.Logger(ctx).Info("nop initializer")
	return nil
}

func finalizeBlock(k *keeper.Keeper, ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	k.Logger(ctx).Info("nop finalizeBlock")
	return nil
}
func commit(k *keeper.Keeper, ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) error {
	k.Logger(ctx).Info("nop commit")
	return nil
}

var Submodule = keeper.Submodule{
	Name:                "nop",
	Version:             "v0.0.1",
	Prepare:             preparer,
	Initialize:          initializer,
	HandleFinalizeBlock: finalizeBlock,
	HandleCommit:        commit,
}
