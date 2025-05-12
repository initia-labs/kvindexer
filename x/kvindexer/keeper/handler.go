package keeper

import (
	"context"
	"fmt"
	"runtime/debug"

	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/initia-labs/kvindexer/x/kvindexer/types"
	"github.com/pkg/errors"
)

func (k *Keeper) Prepare(ctxMap map[string]context.Context) (err error) {
	if !k.config.IsEnabled() {
		return nil
	}

	for _, svc := range k.submodules {
		if err = svc.Prepare(ctxMap[svc.Name()]); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to prepare submodule %s", svc.Name()))
		}
	}

	return nil
}

func (k *Keeper) Start(ctxMap map[string]context.Context) (err error) {
	if !k.config.IsEnabled() {
		return nil
	}

	for _, svc := range k.submodules {
		if err = svc.Initialize(ctxMap[svc.Name()]); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to initialize submodule %s", svc.Name()))
		}
	}

	return nil
}

func (k Keeper) Validate() error {
	if k.config.IsEnabled() {
		if k.db == nil {
			return fmt.Errorf("db is nil")
		}
	}
	// NOP for now
	return nil
}

func (k *Keeper) RegisterSubmodules(submodules ...types.Submodule) error {
	if !k.config.IsEnabled() {
		return nil
	}

	registeredNames := map[string]bool{}

	for _, registered := range submodules {
		if registered.Name() == "" {
			return fmt.Errorf("submodule name must be set")
		}
		if registered.Version() == "" {
			return fmt.Errorf("submodule version must be set")
		}
		if _, found := registeredNames[registered.Name()]; found {
			return fmt.Errorf("submodule %s is duplicated", registered.Name())
		}
		registeredNames[registered.Name()] = true

		k.submodules = append(k.submodules, registered)
	}

	return nil
}

// HandleFinalizeBlock processes the FinalizeBlock event for all submodules.
func (k *Keeper) HandleFinalizeBlock(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) (err error) {
	if !k.config.IsEnabled() {
		return nil
	}

	defer func() {
		if err := recover(); err != nil {
			k.Logger(ctx).Error("panic in HandleFinalizeBlock", "err", err)
			debug.PrintStack()
		}
	}()

	for name, svc := range k.submodules {
		if err = svc.FinalizeBlock(ctx, req, res); err != nil {
			k.Logger(ctx).Warn("failed to handle finalize block event", "submodule", name)
		}
	}

	// pruning
	if k.config.RetainHeight > 0 {
		k.prune(ctx, req.Height)
	}

	return nil
}

func (k *Keeper) HandleCommit(ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) (err error) {
	if !k.config.IsEnabled() {
		return nil
	}

	defer func() {
		if err := recover(); err != nil {
			k.Logger(ctx).Error("panic in HandleCommit", "err", err)
			debug.PrintStack()
		}
	}()

	for name, svc := range k.submodules {
		if err := svc.Commit(ctx, res, changeSet); err != nil {
			k.Logger(ctx).Warn("failed to handle commit event", "submodule", name)
		}
	}

	k.store.Write()

	return nil
}

func (k *Keeper) prune(ctx context.Context, height int64) {

	if running := k.pruningRunning.Swap(true); running {
		return
	}

	go func(ctx context.Context, height int64) {
		defer k.pruningRunning.Store(false)

		minHeight := height - k.config.RetainHeight
		if minHeight <= 0 || minHeight >= height {
			return
		}

		for _, svc := range k.submodules {
			if err := svc.Prune(ctx, minHeight); err != nil {
				k.Logger(ctx).Error("failed to prune", "name", svc.Name(), "error", err)
			}
		}
	}(ctx, height)
}
