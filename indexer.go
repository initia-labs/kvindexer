package indexer

import (
	"context"
	"errors"
	"os"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/initia-labs/kvindexer/config"
	"github.com/initia-labs/kvindexer/module/keeper"
)

var _ storetypes.ABCIListener = Indexer{}

// If indexer.enable is false, it returns (nil, nil)
func NewIndexer(logger log.Logger, k *keeper.Keeper) (*Indexer, error) {
	logger = logger.With("module", "indexer")
	c := k.GetConfig()
	if !c.Enable {
		return nil, nil
	}

	if err := k.Validate(); err != nil {
		return nil, err
	}

	indexer := Indexer{
		config: k.GetConfig(),
		keeper: k,
		logger: logger,
	}

	return &indexer, nil
}

func (i Indexer) Prepare(ctxMap map[string]context.Context) error {
	if !(i.config.Enable) {
		i.logger.Info("indexer is disabled: it won't start.")
		return nil
	}
	return i.keeper.Prepare(ctxMap)
}

func (i Indexer) Start(ctxMap map[string]context.Context) error {
	if !(i.config.Enable) {
		i.logger.Info("indexer is disabled: it won't start.")
		return nil
	}
	if !i.keeper.IsSealed() {
		return errors.New("indexer cannot start because the keeper is not sealed")
	}
	return i.keeper.Start(ctxMap)
}

func (i Indexer) Validate() error {
	if !(i.config.Enable) {
		i.logger.Debug("indexer is disabled: no validation needed.")
		return nil
	}
	if err := i.config.Validate(); err != nil {
		return err
	}
	if err := i.keeper.Validate(); err != nil {
		return err
	}
	return nil
}

// It opens a batch before handling FinalizeBlock
func (i Indexer) ListenFinalizeBlock(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	ctx, _ = sdkCtx.CacheContext()

	err := i.keeper.HandleFinalizeBlock(ctx, req, res)
	if err != nil {
		i.logger.Error("failed to handle finalize block", "err", err)
		os.Exit(1)
	}
	return err
}

// and It closes the batch after handling Commit.
func (i Indexer) ListenCommit(ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	ctx, _ = sdkCtx.CacheContext()

	err := i.keeper.HandleCommit(ctx, res, changeSet)
	if err != nil {
		i.logger.Error("failed to handle commit", "err", err)
		os.Exit(1)
	}
	err = i.keeper.WriteStore()

	return err
}

func PreSetupIndexer(svrCtx *server.Context, clientCtx client.Context, ctx context.Context, g *errgroup.Group, _app servertypes.Application) error {
	app, ok := _app.(IndexableApplication)
	if !ok {
		return errors.New("app is not IndexableApplication")
	}

	// listen all keys
	keysToListen := []storetypes.StoreKey{}
	keys := app.GetKeys()
	for _, key := range keys {
		keysToListen = append(keysToListen, key)
	}
	app.GetBaseApp().CommitMultiStore().AddListeners(keysToListen)

	indexer, err := NewIndexer(app.GetBaseApp().Logger(), app.GetIndexerKeeper())
	// if err is not nil, it means there is an error regardless of indexer is nil or not.
	// else if indexer is nil, it means indexer is disabled and the returned err is nil.
	if err != nil || indexer == nil {
		return err
	}

	if err = indexer.Validate(); err != nil {
		return err
	}

	if err = indexer.Prepare(nil); err != nil {
		return err
	}

	if err = app.GetIndexerKeeper().Seal(); err != nil {
		return err
	}

	if err = indexer.Start(nil); err != nil {
		return err
	}

	streamingManager := storetypes.StreamingManager{
		ABCIListeners: []storetypes.ABCIListener{indexer},
		StopNodeOnErr: true,
	}
	app.GetBaseApp().SetStreamingManager(streamingManager)

	return nil
}

func AddIndexFlag(cmd *cobra.Command) {
	config.AddIndexerFlag(cmd)
}

var StartCmdOptions = server.StartCmdOptions{
	DBOpener: nil,
	PreSetup: PreSetupIndexer,
	AddFlags: AddIndexFlag,
}
