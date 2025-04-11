package tx

import (
	"context"
	"sync"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/initia-labs/kvindexer/collection"
	"github.com/initia-labs/kvindexer/submodules/evm-tx/types"

	"github.com/pkg/errors"
)

const (
	oldModuleName = "tx"
)

var runOnce sync.Once

func (s *EvmTxSubmodule) PatchPrefix(ctx context.Context) (err error) {
	runOnce.Do(func() {
		s.Logger(ctx).Info("patching EVM-TX submodule prefix...")
		err = s.patchPrefix(ctx)
		s.Logger(ctx).Info("patch EVM-TX submodule prefix done", "err", err)
	})

	s.Logger(ctx).Info("invoked goruoutine to patch EVM-TX submodule prefix...")
	return err
}

func (s *EvmTxSubmodule) patchPrefix(ctx context.Context) (err error) {

	wg := sync.WaitGroup{}

	// patch sequence
	wg.Add(1)
	go func() {
		err = s.patchSequence(ctx)
		if err != nil {
			s.Logger(ctx).Error("failed to patch sequence", "err", err)
		} else {
			s.Logger(ctx).Info("successfully patched sequence")
		}
		wg.Done()
	}()

	// patch sequence
	wg.Add(1)
	go func() {
		err = s.patchTxMap(ctx)
		if err != nil {
			s.Logger(ctx).Error("failed to patch sequence", "err", err)
		} else {
			s.Logger(ctx).Info("successfully patched sequence")
		}
		wg.Done()
	}()

	// patch sequence
	wg.Add(1)
	go func() {
		err = s.patchTxhashesByAccountMap(ctx)
		if err != nil {
			s.Logger(ctx).Error("failed to patch sequence", "err", err)
		} else {
			s.Logger(ctx).Info("successfully patched sequence")
		}
		wg.Done()
	}()

	// patch sequence
	wg.Add(1)
	go func() {
		err = s.patchTxhashesBySequenceMap(ctx)
		if err != nil {
			s.Logger(ctx).Error("failed to patch sequence", "err", err)
		} else {
			s.Logger(ctx).Info("successfully patched sequence")
		}
		wg.Done()
	}()

	// patch sequence
	wg.Add(1)
	go func() {
		err = s.patcAccountSequenceMap(ctx)
		if err != nil {
			s.Logger(ctx).Error("failed to patch sequence", "err", err)
		} else {
			s.Logger(ctx).Info("successfully patched sequence")
		}
		wg.Done()
	}()

	// patch sequence
	wg.Add(1)
	go func() {
		err = s.patchSequenceByHeightMap(ctx)
		if err != nil {
			s.Logger(ctx).Error("failed to patch sequence", "err", err)
		} else {
			s.Logger(ctx).Info("successfully patched sequence")
		}
		wg.Done()
	}()

	// patch sequence
	wg.Add(1)
	go func() {
		err = s.patchAccountSequenceByHeightMap(ctx)
		if err != nil {
			s.Logger(ctx).Error("failed to patch sequence", "err", err)
		} else {
			s.Logger(ctx).Info("successfully patched sequence")
		}
		wg.Done()
	}()

	wg.Wait()

	return err
}

func (s *EvmTxSubmodule) patchSequence(ctx context.Context) (err error) {
	oldPrefix := collection.NewPrefix(oldModuleName, types.SequencePrefix)
	oldSeq, err := collection.AddSequence(s.keeper, oldPrefix, "sequence")
	if err != nil {
		return errors.Wrap(err, "failed to get old sequence")
	}

	oldval, err := oldSeq.Peek(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get old sequence value")
	}
	curval, err := s.sequence.Peek(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get current sequence value")
	}

	err = s.sequence.Set(ctx, curval+oldval)
	if err != nil {
		return errors.Wrap(err, "failed to set current sequence value")
	}

	return nil
}

func (s *EvmTxSubmodule) patchTxMap(ctx context.Context) (err error) {
	oldPrefix := collection.NewPrefix(oldModuleName, types.TxsPrefix)
	oldTxMap, err := collection.AddMap(s.keeper, oldPrefix, "txs", collections.StringKey, codec.CollValue[sdk.TxResponse](s.cdc))
	if err != nil {
		return errors.Wrap(err, "failed to get old tx map")
	}

	err = oldTxMap.Walk(ctx, nil, func(key string, value sdk.TxResponse) (stop bool, err error) {
		err = s.txMap.Set(ctx, key, value)
		if err == nil {
			return true, errors.Wrap(err, "failed to set tx map")
		}

		// remove old key
		err = oldTxMap.Remove(ctx, key)
		if err != nil {
			return true, errors.Wrap(err, "failed to remove old tx map")
		}
		return false, nil

	})
	if err != nil {
		return errors.Wrap(err, "failed to walk through old tx map")
	}

	return nil
}

func (s *EvmTxSubmodule) patchTxhashesByAccountMap(ctx context.Context) (err error) {
	oldPrefix := collection.NewPrefix(oldModuleName, types.TxsByAccountPrefix)
	oldTxhashesByAccountMap, err := collection.AddMap(s.keeper, oldPrefix, "txs_by_account", collections.PairKeyCodec(sdk.AccAddressKey, collections.Uint64Key), collections.StringValue)
	if err != nil {
		return errors.Wrap(err, "failed to get old txhashedByAccount map")
	}

	err = oldTxhashesByAccountMap.Walk(ctx, nil, func(key collections.Pair[sdk.AccAddress, uint64], value string) (stop bool, err error) {
		err = s.txhashesByAccountMap.Set(ctx, key, value)
		if err == nil {
			return true, errors.Wrap(err, "failed to set txhashedByAccount map")
		}

		// remove old key
		err = oldTxhashesByAccountMap.Remove(ctx, key)
		if err != nil {
			return true, errors.Wrap(err, "failed to remove old txhashedByAccount map")
		}
		return false, nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to walk through old txhashedByAccount map")
	}

	return nil
}

func (s *EvmTxSubmodule) patchTxhashesBySequenceMap(ctx context.Context) (err error) {
	oldPRefix := collection.NewPrefix(oldModuleName, types.TxSequencePrefix)
	oldTxhashesBySequenceMap, err := collection.AddMap(s.keeper, oldPRefix, "tx_sequences", collections.Uint64Key, collections.StringValue)
	if err != nil {
		return errors.Wrap(err, "failed to get old txhashesBySequence map")
	}

	err = oldTxhashesBySequenceMap.Walk(ctx, nil, func(key uint64, value string) (stop bool, err error) {
		err = s.txhashesBySequenceMap.Set(ctx, key, value)
		if err == nil {
			return true, errors.Wrap(err, "failed to set txhashesBySequence map")
		}

		// remove old key
		err = oldTxhashesBySequenceMap.Remove(ctx, key)
		if err != nil {
			return true, errors.Wrap(err, "failed to remove old txhashesBySequence map")
		}
		return false, nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to walk through old txhashesBySequence map")
	}

	return nil
}

func (s *EvmTxSubmodule) patchTxhashesByHeightMap(ctx context.Context) (err error) {
	oldPrefix := collection.NewPrefix(oldModuleName, types.TxByHeightPrefix)
	oldTxhashesByHeightMap, err := collection.AddMap(s.keeper, oldPrefix, "txs_by_height", collections.PairKeyCodec(collections.Int64Key, collections.Uint64Key), collections.StringValue)
	if err != nil {
		return errors.Wrap(err, "failed to get old txhashesByHeight map")
	}

	err = oldTxhashesByHeightMap.Walk(ctx, nil, func(key collections.Pair[int64, uint64], value string) (stop bool, err error) {
		err = s.txhashesByHeightMap.Set(ctx, key, value)
		if err == nil {
			return true, errors.Wrap(err, "failed to set txhashesByHeight map")
		}

		// remove old key
		err = oldTxhashesByHeightMap.Remove(ctx, key)
		if err != nil {
			return true, errors.Wrap(err, "failed to remove old txhashesByHeight map")
		}
		return false, nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to walk through old txhashesByHeight map")
	}

	return nil
}

func (s *EvmTxSubmodule) patcAccountSequenceMap(ctx context.Context) (err error) {
	oldPrefix := collection.NewPrefix(oldModuleName, types.AccountSequencePrefix)
	oldAccountSequenceMap, err := collection.AddMap(s.keeper, oldPrefix, "account_sequences", sdk.AccAddressKey, collections.Uint64Value)
	if err != nil {
		return errors.Wrap(err, "failed to get old accountSequence map")
	}

	err = oldAccountSequenceMap.Walk(ctx, nil, func(key sdk.AccAddress, value uint64) (stop bool, err error) {
		err = s.accountSequenceMap.Set(ctx, key, value)
		if err == nil {
			return true, errors.Wrap(err, "failed to set accountSequence map")
		}

		// remove old key
		err = oldAccountSequenceMap.Remove(ctx, key)
		if err != nil {
			return true, errors.Wrap(err, "failed to remove old accountSequence map")
		}
		return false, nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to walk through old accountSequence map")
	}

	return nil
}

func (s *EvmTxSubmodule) patchSequenceByHeightMap(ctx context.Context) (err error) {
	oldPrefix := collection.NewPrefix(oldModuleName, types.SequenceByHeightPrefix)
	oldSequenceByHeightMap, err := collection.AddMap(s.keeper, oldPrefix, "sequence_by_height", collections.Int64Key, collections.Uint64Value)
	if err != nil {
		return errors.Wrap(err, "failed to get old sequenceByHeight map")
	}

	err = oldSequenceByHeightMap.Walk(ctx, nil, func(key int64, value uint64) (stop bool, err error) {
		err = s.sequenceByHeightMap.Set(ctx, key, value)
		if err == nil {
			return true, errors.Wrap(err, "failed to set sequenceByHeight map")
		}

		// remove old key
		err = oldSequenceByHeightMap.Remove(ctx, key)
		if err != nil {
			return true, errors.Wrap(err, "failed to remove old sequenceByHeight map")
		}
		return false, nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to walk through old sequenceByHeight map")
	}

	return nil
}

func (s *EvmTxSubmodule) patchAccountSequenceByHeightMap(ctx context.Context) (err error) {
	oldPrefix := collection.NewPrefix(oldModuleName, types.AccountSequenceByHeightPrefix)
	oldAccountSequenceByHeightMap, err := collection.AddMap(s.keeper, oldPrefix, "account_sequence_by_height", collections.TripleKeyCodec(collections.Int64Key, sdk.AccAddressKey, collections.Uint64Key), collections.BoolValue)
	if err != nil {
		return errors.Wrap(err, "failed to get old accountSequenceByHeight map")
	}

	err = oldAccountSequenceByHeightMap.Walk(ctx, nil, func(key collections.Triple[int64, sdk.AccAddress, uint64], value bool) (stop bool, err error) {
		err = s.accountSequenceByHeightMap.Set(ctx, key, value)
		if err == nil {
			return true, errors.Wrap(err, "failed to set accountSequenceByHeight map")
		}

		// remove old key
		err = oldAccountSequenceByHeightMap.Remove(ctx, key)
		if err != nil {
			return true, errors.Wrap(err, "failed to remove old accountSequenceByHeight map")
		}
		return false, nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to walk through old accountSequenceByHeight map")
	}

	return nil
}
