package tx

import (
	"context"
	"sync"
	"time"

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
		start := time.Now()
		err = s.patchPrefix(ctx)
		done := time.Since(start)
		s.Logger(ctx).Info("patch EVM-TX submodule prefix done", "err", err, "took", done.Seconds())
	})

	return err
}

func (s *EvmTxSubmodule) patchPrefix(ctx context.Context) (err error) {

	wg := sync.WaitGroup{}

	oldSeq, err := s.patchSequence(ctx)
	if err != nil {
		s.Logger(ctx).Error("failed to patch sequence", "err", err)
	} else {
		s.Logger(ctx).Info("successfully patched sequence", "oldSeq", oldSeq)
	}

	wg.Add(1)
	go func() {
		err = s.patchTxMap(ctx)
		if err != nil {
			s.Logger(ctx).Error("failed to patch txMap", "err", err)
		} else {
			s.Logger(ctx).Info("successfully patched txMap")
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		err = s.patcAccountSequenceMap(ctx)
		if err != nil {
			s.Logger(ctx).Error("failed to patch AccountSequenceMap", "err", err)
		} else {
			s.Logger(ctx).Info("successfully patched AccountSequenceMap")
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		err = s.patchTxhashesByAccountMap(ctx)
		if err != nil {
			s.Logger(ctx).Error("failed to patch TxhashesByAccountMap", "err", err)
		} else {
			s.Logger(ctx).Info("successfully patched TxhashesByAccountMap")
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		err = s.patchTxhashesBySequenceMap(ctx, oldSeq)
		if err != nil {
			s.Logger(ctx).Error("failed to patch TxhashesBySequenceMap", "err", err)
		} else {
			s.Logger(ctx).Info("successfully patched TxhashesBySequenceMap")
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		err = s.patchSequenceByHeightMap(ctx, oldSeq)
		if err != nil {
			s.Logger(ctx).Error("failed to patch SequenceByHeightMap", "err", err)
		} else {
			s.Logger(ctx).Info("successfully patched SequenceByHeightMap")
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		err = s.patchTxhashesByHeightMap(ctx)
		if err != nil {
			s.Logger(ctx).Error("failed to patch TxhashesByHeightMap", "err", err)
		} else {
			s.Logger(ctx).Info("successfully patched TxhashesByHeightMap")
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		err = s.patchAccountSequenceByHeightMap(ctx, oldSeq)
		if err != nil {
			s.Logger(ctx).Error("failed to patch AccountSequenceByHeightMap", "err", err)
		} else {
			s.Logger(ctx).Info("successfully patched AccountSequenceByHeightMap")
		}
		wg.Done()
	}()

	wg.Wait()

	return err
}

func (s *EvmTxSubmodule) patchSequence(ctx context.Context) (lastSeq uint64, err error) {
	oldPrefix := collection.NewPrefix(oldModuleName, types.SequencePrefix)
	oldSeq, err := collection.AddSequence(s.keeper, oldPrefix, "sequence")
	if err != nil {
		return 0, errors.Wrap(err, "failed to get old sequence")
	}

	oldval, err := oldSeq.Peek(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get old sequence value")
	}
	curval, err := s.sequence.Peek(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get current sequence value")
	}

	err = s.sequence.Set(ctx, curval+oldval)
	if err != nil {
		return 0, errors.Wrap(err, "failed to set current sequence value")
	}

	return oldval, nil
}

func (s *EvmTxSubmodule) patchTxMap(ctx context.Context) (err error) {
	oldPrefix := collection.NewPrefix(oldModuleName, types.TxsPrefix)
	oldTxMap, err := collection.AddMap(s.keeper, oldPrefix, "txs", collections.StringKey, codec.CollValue[sdk.TxResponse](s.cdc))
	if err != nil {
		return errors.Wrap(err, "failed to get old tx map")
	}

	// no need to patch current txMap
	i := 0
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

		i++
		if i%1000 == 0 {
			s.Logger(ctx).Info("patching tx map", "count", i)
		}

		return false, nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to walk through old tx map")
	}

	return nil
}

func (s *EvmTxSubmodule) patcAccountSequenceMap(ctx context.Context) (err error) {
	oldPrefix := collection.NewPrefix(oldModuleName, types.AccountSequencePrefix)
	oldAccountSequenceMap, err := collection.AddMap(s.keeper, oldPrefix, "account_sequences", sdk.AccAddressKey, collections.Uint64Value)
	if err != nil {
		return errors.Wrap(err, "failed to get old accountSequence map")
	}

	i := 0
	err = oldAccountSequenceMap.Walk(ctx, nil, func(key sdk.AccAddress, oldVal uint64) (stop bool, err error) {

		var curVal uint64
		curVal, err = s.accountSequenceMap.Get(ctx, key)
		if err != nil {
			return true, errors.Wrap(err, "failed to get current accountSequence map")
		}
		err = s.accountSequenceMap.Set(ctx, key, curVal+oldVal)
		if err == nil {
			return true, errors.Wrap(err, "failed to set accountSequence map")
		}

		// remove old key
		err = oldAccountSequenceMap.Remove(ctx, key)
		if err != nil {
			return true, errors.Wrap(err, "failed to remove old accountSequence map")
		}

		i++
		if i%1000 == 0 {
			s.Logger(ctx).Info("patching account sequence map", "count", i)
		}
		return false, nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to walk through old accountSequence map")
	}

	return nil
}

func (s *EvmTxSubmodule) patchTxhashesByAccountMap(ctx context.Context) (err error) {
	oldPrefix := collection.NewPrefix(oldModuleName, types.TxsByAccountPrefix)
	oldTxhashesByAccountMap, err := collection.AddMap(s.keeper, oldPrefix, "txs_by_account", collections.PairKeyCodec(sdk.AccAddressKey, collections.Uint64Key), collections.StringValue)
	if err != nil {
		return errors.Wrap(err, "failed to get old txhashedByAccount map")
	}
	// just extracted map from current store
	curMap := make(map[collections.Pair[sdk.AccAddress, uint64]]string)
	// key: address, value: last seq from old store
	lastSeq := make(map[string]uint64)

	i := 0
	err = s.txhashesByAccountMap.Walk(ctx, nil, func(key collections.Pair[sdk.AccAddress, uint64], value string) (stop bool, err error) {
		curMap[key] = value
		err = s.txhashesByAccountMap.Remove(ctx, key)
		i++
		if i%1000 == 0 {
			s.Logger(ctx).Info("patching txhashedByAccount map", "count", i, "step", "remove from current store")
		}
		return err != nil, errors.Wrap(err, "failed to pop prev txhashedByAccount map")
	})
	if err != nil {
		return errors.Wrap(err, "failed to walk through cur txhashedByAccount map")
	}

	i = 0
	err = oldTxhashesByAccountMap.Walk(ctx, nil, func(key collections.Pair[sdk.AccAddress, uint64], value string) (stop bool, err error) {

		// get last seq from old store
		lseq, ok := lastSeq[key.K1().String()]
		if !ok {
			lseq = key.K2()
		} else {
			if lseq < key.K2() {
				lastSeq[key.K1().String()] = key.K2()
			}
		}

		err = s.txhashesByAccountMap.Set(ctx, key, value)
		if err == nil {
			return true, errors.Wrap(err, "failed to set txhashedByAccount map")
		}

		// remove old key
		err = oldTxhashesByAccountMap.Remove(ctx, key)
		if err != nil {
			return true, errors.Wrap(err, "failed to remove old txhashedByAccount map")
		}
		i++
		if i%1000 == 0 {
			s.Logger(ctx).Info("patching txhashedByAccount map", "count", i, "step", "migrate old store")
		}
		return false, nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to walk through old txhashedByAccount map")
	}

	// re-insert curmap
	i = 0
	for k, v := range curMap {
		err = s.txhashesByAccountMap.Set(ctx, collections.Join(k.K1(), lastSeq[k.K1().String()]), v)
		if err != nil {
			return errors.Wrap(err, "failed to set txhashedByAccount map")
		}
		i++
		if i%1000 == 0 {
			s.Logger(ctx).Info("patching txhashedByAccount map", "count", i, "step", "re-insert cur store")
		}
	}

	return nil
}

func (s *EvmTxSubmodule) patchTxhashesBySequenceMap(ctx context.Context, oldSeq uint64) (err error) {
	oldPRefix := collection.NewPrefix(oldModuleName, types.TxSequencePrefix)
	oldTxhashesBySequenceMap, err := collection.AddMap(s.keeper, oldPRefix, "tx_sequences", collections.Uint64Key, collections.StringValue)
	if err != nil {
		return errors.Wrap(err, "failed to get old txhashesBySequence map")
	}

	i := 0
	// patch current txhashesBySequenceMap
	err = s.txhashesBySequenceMap.Walk(ctx, nil, func(key uint64, value string) (stop bool, err error) {
		err = s.txhashesBySequenceMap.Remove(ctx, key)
		if err != nil {
			return true, errors.Wrap(err, "failed to remove prev txhashesBySequence map")
		}
		err = s.txhashesBySequenceMap.Set(ctx, key+oldSeq, value)
		if err != nil {
			return true, errors.Wrap(err, "failed to set txhashesBySequence map")
		}
		i++
		if i%1000 == 0 {
			s.Logger(ctx).Info("patching txhashesBySequence map", "count", i, "step", "update current store")
		}
		return false, nil
	})

	i = 0
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
		i++
		if i%1000 == 0 {
			s.Logger(ctx).Info("patching txhashesBySequence map", "count", i, "step", "migrate old store")
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

	i := 0
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
		i++
		if i%1000 == 0 {
			s.Logger(ctx).Info("patching txhashesByHeight map", "count", i, "step", "migrate old store")
		}
		return false, nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to walk through old txhashesByHeight map")
	}

	return nil
}

func (s *EvmTxSubmodule) patchSequenceByHeightMap(ctx context.Context, oldSeq uint64) (err error) {
	oldPrefix := collection.NewPrefix(oldModuleName, types.SequenceByHeightPrefix)
	oldSequenceByHeightMap, err := collection.AddMap(s.keeper, oldPrefix, "sequence_by_height", collections.Int64Key, collections.Uint64Value)
	if err != nil {
		return errors.Wrap(err, "failed to get old sequenceByHeight map")
	}

	i := 0
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
		i++
		if i%1000 == 0 {
			s.Logger(ctx).Info("patching sequenceByHeight map", "count", i, "step", "migrate old store")
		}
		return false, nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to walk through old sequenceByHeight map")
	}

	return nil
}

func (s *EvmTxSubmodule) patchAccountSequenceByHeightMap(ctx context.Context, oldSeq uint64) (err error) {
	oldPrefix := collection.NewPrefix(oldModuleName, types.AccountSequenceByHeightPrefix)
	oldAccountSequenceByHeightMap, err := collection.AddMap(s.keeper, oldPrefix, "account_sequence_by_height", collections.TripleKeyCodec(collections.Int64Key, sdk.AccAddressKey, collections.Uint64Key), collections.BoolValue)
	if err != nil {
		return errors.Wrap(err, "failed to get old accountSequenceByHeight map")
	}

	i := 0
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
		i++
		if i%1000 == 0 {
			s.Logger(ctx).Info("patching accountSequenceByHeight map", "count", i, "step", "migrate old store")
		}

		return false, nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to walk through old accountSequenceByHeight map")
	}

	return nil
}
