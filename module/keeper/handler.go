package keeper

import (
	"context"
	"fmt"
	"strings"

	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/pkg/errors"
)

func (k *Keeper) Prepare(ctxMap map[string]context.Context) (err error) {
	for name, svc := range k.submodules {
		if svc.Prepare != nil {
			if !k.config.IsEnabledSubmodule(name) {
				continue
			}
			fn := svc.Prepare
			if fn == nil {
				continue
			}
			if err = (fn)(k, ctxMap[name], k.config.SubmoduleConfigs[name]); err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to prepare submodule %s", name))
			}
		}
	}

	return nil
}

func (k *Keeper) Start(ctxMap map[string]context.Context) (err error) {
	if err = k.crontab.Initialize(); err != nil {
		return errors.Wrap(err, "failed to initialize crontab")
	}
	k.crontab.Start()

	for name, svc := range k.submodules {
		if svc.Initialize != nil {
			if !k.config.IsEnabledSubmodule(name) {
				continue
			}
			if err = (svc.Initialize)(k, ctxMap[name], k.config.SubmoduleConfigs[name]); err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to initialize submodule %s", name))
			}
		}
	}

	return nil
}

func (k Keeper) Validate() error {
	// NOP for now
	return nil
}

func (k *Keeper) RegisterSubmodules(svcs ...Submodule) error {
	if !k.config.IsEnabled() {
		return nil
	}

	for _, svc := range svcs {
		if _, found := k.submodules[svc.Name]; found {
			return fmt.Errorf("submodule %s is duplicated", svc.Name)
		}

		for prevName := range k.submodules {
			if strings.HasPrefix(prevName, svc.Name) || strings.HasPrefix(svc.Name, prevName) {
				return fmt.Errorf("submodule %s is overlapping with %s", svc.Name, prevName)
			}
		}

		k.submodules[svc.Name] = svc
	}

	return nil
}

func (k *Keeper) HandleFinalizeBlock(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) (err error) {
	for name, svc := range k.submodules {
		if !k.config.IsEnabledSubmodule(name) {
			continue
		}
		fn := svc.HandleFinalizeBlock
		if fn == nil {
			continue
		}

		if err = (fn)(k, ctx, req, res, k.config.SubmoduleConfigs[name]); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to handle finalize block event for submodule %s", name))
		}
	}
	return nil
}

func (k *Keeper) HandleCommit(ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) (err error) {
	for name, svc := range k.submodules {
		if !k.config.IsEnabledSubmodule(name) {
			continue
		}
		fn := svc.HandleCommit
		if fn == nil {
			continue
		}
		if err := (fn)(k, ctx, res, changeSet, k.config.SubmoduleConfigs[name]); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to handle commit event for submodule %s", name))
		}
	}
	return nil
}
