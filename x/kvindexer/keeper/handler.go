package keeper

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"

	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/initia-labs/kvindexer/x/kvindexer/types"
	"github.com/pkg/errors"
)

func (k *Keeper) Prepare(ctxMap map[string]context.Context) (err error) {
	for name, svc := range k.submodules {
		if err = svc.Prepare(ctxMap[name]); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to prepare submodule %s", name))
		}
	}

	return nil
}

func (k *Keeper) Start(ctxMap map[string]context.Context) (err error) {
	for name, svc := range k.submodules {
		if err = svc.Initialize(ctxMap[name]); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to initialize submodule %s", name))
		}
	}

	return nil
}

func (k Keeper) Validate() error {
	// NOP for now
	return nil
}

func (k *Keeper) RegisterSubmodules(submodules ...types.Submodule) error {
	if !k.config.IsEnabled() {
		return nil
	}

	for _, submodule := range submodules {
		if submodule.Name() == "" {
			return fmt.Errorf("submodule name must be set")
		}
		if submodule.Version() == "" {
			return fmt.Errorf("submodule version must be set")
		}
		if _, found := k.submodules[submodule.Name()]; found {
			return fmt.Errorf("submodule %s is duplicated", submodule.Name())
		}

		for prevName := range k.submodules {
			if strings.HasPrefix(prevName, submodule.Name()) || strings.HasPrefix(submodule.Name(), prevName) {
				return fmt.Errorf("submodule %s is overlapping with %s", submodule.Name(), prevName)
			}
		}

		k.submodules[submodule.Name()] = submodule
	}

	return nil
}

func (k *Keeper) HandleFinalizeBlock(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) (err error) {
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

	return nil
}

func (k *Keeper) HandleCommit(ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) (err error) {
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
