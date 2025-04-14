package tx

import (
	"context"
	"sync"
	"time"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/pkg/errors"
)

const remove = true
const inSync = true
const iterUnit = 10_000

const (
	oldModuleName = "tx"
)

var runOnce sync.Once

func (s *EvmTxSubmodule) PatchPrefix(ctx context.Context, do bool) (err error) {
	if !do {
		return nil
	}
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
	if inSync {
		wg.Wait()
	}

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
	if inSync {
		wg.Wait()
	}

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

	if inSync {
		wg.Wait()
	}
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

	if inSync {
		wg.Wait()
	}

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
	// no inSync wait here
	/**
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
	      err = s.patchAccountSequenceByHeightMap(ctx, oldSeq)
	      if err != nil {
	          s.Logger(ctx).Error("failed to patch AccountSequenceByHeightMap", "err", err)
	      } else {
	          s.Logger(ctx).Info("successfully patched AccountSequenceByHeightMap")
	      }
	      wg.Done()
	  }()
	*/

	wg.Wait()

	return err
}

func (s *EvmTxSubmodule) patchSequence(ctx context.Context) (lastSeq uint64, err error) {

	oldval, err := s.oldSequence.Peek(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get old sequence value")
	}
	curval, err := s.sequence.Peek(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get current sequence value")
	}

	s.Logger(ctx).Info("patching sequence", "oldval", oldval, "curval", curval)

	err = s.sequence.Set(ctx, curval+oldval)
	if err != nil {
		return 0, errors.Wrap(err, "failed to set current sequence value")
	}

	return oldval, nil
}

func (s *EvmTxSubmodule) patchTxMap(ctx context.Context) (err error) {

	// no need to patch current txMap
	i := 0
	pageReq := &query.PageRequest{
		Key:        nil,
		Offset:     0,
		Limit:      iterUnit,
		CountTotal: false,
	}
	for {
		txMap := make(map[string]sdk.TxResponse)
		isFirst := true
		_, pageRes, err := query.CollectionPaginate(ctx, s.oldTxMap, pageReq, func(key string, value sdk.TxResponse) (res sdk.TxResponse, err error) {
			txMap[key] = value

			if isFirst {
				isFirst = false
				s.Logger(ctx).Info("FIRST ITEM", "key", key)
			}

			// remove old key
			if remove {
				err = s.oldTxMap.Remove(ctx, key)
				if err != nil {
					return res, errors.Wrap(err, "failed to remove old tx map")
				}
			}

			return value, nil
		})
		if err != nil {
			return errors.Wrap(err, "failed to walk through old tx map")
		}
		if pageRes == nil || len(pageRes.NextKey) == 0 {
			break
		}
		i += len(txMap)

		s.Logger(ctx).Info("patching tx map", "count", len(txMap), "nextKey", string(pageRes.NextKey))
		//pageReq.Key = pageRes.NextKey // removed to avoid infinite loop

		for k, v := range txMap {
			err = s.txMap.Set(ctx, k, v)
			if err != nil {
				return errors.Wrap(err, "failed to set tx map")
			}
		}
		s.Logger(ctx).Info("migrate tx map", "count", len(txMap))
		clear(txMap)
		s.Logger(ctx).Info("cleared tx map", "count", len(txMap))
		if err = s.keeper.Write(); err != nil {
			return errors.Wrap(err, "failed to write tx map")
		}
		s.Logger(ctx).Info("write tx map", "count", len(txMap))
		//pageReq.Key = pageRes.NextKey
		s.Logger(ctx).Info("patching tx map...", "count", i)
	}
	s.Logger(ctx).Info("patching tx map done", "count", i)
	if err != nil {
		return errors.Wrap(err, "failed to walk through old tx map")
	}

	return nil
}

func (s *EvmTxSubmodule) patcAccountSequenceMap(ctx context.Context) (err error) {

	i := 0
	err = s.oldAccountSequenceMap.Walk(ctx, nil, func(key sdk.AccAddress, oldVal uint64) (stop bool, err error) {

		var curVal uint64
		curVal, err = s.accountSequenceMap.Get(ctx, key)
		if err != nil {
			return true, errors.Wrap(err, "failed to get current accountSequence map")
		}
		err = s.accountSequenceMap.Set(ctx, key, curVal+oldVal)
		if err != nil {
			return true, errors.Wrap(err, "failed to set accountSequence map")
		}

		// remove old key
		if remove {
			err = s.oldAccountSequenceMap.Remove(ctx, key)
			if err != nil {
				return true, errors.Wrap(err, "failed to remove old accountSequence map")
			}
		}

		s.Logger(ctx).Info("patching account sequence map", "key", key, "oldVal", oldVal, "curVal", curVal)

		i++
		if i%1000 == 0 {
			s.Logger(ctx).Info("patching account sequence map", "count", i)
		}
		return false, nil
	})
	s.Logger(ctx).Info("patching account sequence map", "count", i)
	if err != nil {
		return errors.Wrap(err, "failed to walk through old accountSequence map")
	}

	return nil
}

func (s *EvmTxSubmodule) patchTxhashesByAccountMap(ctx context.Context) (err error) {

	// just extracted map from current store
	curMap := make(map[collections.Pair[sdk.AccAddress, uint64]]string)
	// key: address, value: last seq from old store
	lastSeq := make(map[string]uint64)

	i := 0
	err = s.txhashesByAccountMap.Walk(ctx, nil, func(key collections.Pair[sdk.AccAddress, uint64], value string) (stop bool, err error) {
		curMap[key] = value
		if remove {
			err = s.txhashesByAccountMap.Remove(ctx, key)
		}
		i++
		if i%1000 == 0 {
			s.Logger(ctx).Info("patching txhashedByAccount map", "count", i, "step", "remove from current store")
		}
		return err != nil, errors.Wrap(err, "failed to pop prev txhashedByAccount map")
	})
	if err != nil {
		return errors.Wrap(err, "failed to walk through cur txhashedByAccount map")
	}
	s.Logger(ctx).Info("patched txhashedByAccount map", "count", i, "step", "remove from current store")

	i = 0
	err = s.oldTxhashesByAccountMap.Walk(ctx, nil, func(key collections.Pair[sdk.AccAddress, uint64], value string) (stop bool, err error) {

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
		if err != nil {
			return true, errors.Wrap(err, "failed to set txhashedByAccount map")
		}

		s.Logger(ctx).Info("patching txhashedByAccount map", "key", key, "value", value)

		// remove old key
		if remove {
			err = s.oldTxhashesByAccountMap.Remove(ctx, key)
			if err != nil {
				return true, errors.Wrap(err, "failed to remove old txhashedByAccount map")
			}
		}
		i++
		if i%1000 == 0 {
			s.Logger(ctx).Info("patching txhashedByAccount map", "count", i, "step", "migrate old store")
		}
		return false, nil
	})
	s.Logger(ctx).Info("patched txhashedByAccount map", "count", i, "step", "migrate old store")
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
		s.Logger(ctx).Info("patching txhashedByAccount map", "key", k, "value", v)
		i++
		if i%1000 == 0 {
			s.Logger(ctx).Info("patching txhashedByAccount map", "count", i, "step", "re-insert cur store")
		}
	}
	s.Logger(ctx).Info("patched txhashedByAccount map", "count", i, "step", "re-insert cur store")

	return nil
}

func (s *EvmTxSubmodule) patchTxhashesBySequenceMap(ctx context.Context, oldSeq uint64) (err error) {

	i := 0
	// patch current txhashesBySequenceMap
	err = s.txhashesBySequenceMap.Walk(ctx, nil, func(key uint64, value string) (stop bool, err error) {
		if remove {
			err = s.txhashesBySequenceMap.Remove(ctx, key)
			if err != nil {
				return true, errors.Wrap(err, "failed to remove prev txhashesBySequence map")
			}
		}
		err = s.txhashesBySequenceMap.Set(ctx, key+oldSeq, value)
		if err != nil {
			return true, errors.Wrap(err, "failed to set txhashesBySequence map")
		}
		s.Logger(ctx).Info("patching txhashesBySequence map", "key", key, "value", value)
		i++
		if i%1000 == 0 {
			s.Logger(ctx).Info("patching txhashesBySequence map", "count", i, "step", "update current store")
		}
		return false, nil
	})
	s.Logger(ctx).Info("patched txhashesBySequence map", "count", i, "step", "update current store")

	i = 0
	err = s.oldTxhashesBySequenceMap.Walk(ctx, nil, func(key uint64, value string) (stop bool, err error) {
		err = s.txhashesBySequenceMap.Set(ctx, key, value)
		if err != nil {
			return true, errors.Wrap(err, "failed to set txhashesBySequence map")
		}

		// remove old key
		if remove {
			err = s.oldTxhashesBySequenceMap.Remove(ctx, key)
			if err != nil {
				return true, errors.Wrap(err, "failed to remove old txhashesBySequence map")
			}
		}
		s.Logger(ctx).Info("patching txhashesBySequence map", "key", key, "value", value)
		i++
		if i%1000 == 0 {
			s.Logger(ctx).Info("patching txhashesBySequence map", "count", i, "step", "migrate old store")
		}
		return false, nil
	})
	s.Logger(ctx).Info("patched txhashesBySequence map", "count", i, "step", "migrate old store")
	if err != nil {
		return errors.Wrap(err, "failed to walk through old txhashesBySequence map")
	}

	return nil
}

func (s *EvmTxSubmodule) patchTxhashesByHeightMap(ctx context.Context) (err error) {

	i := 0
	err = s.oldTxhashesByHeightMap.Walk(ctx, nil, func(key collections.Pair[int64, uint64], value string) (stop bool, err error) {
		err = s.txhashesByHeightMap.Set(ctx, key, value)
		if err != nil {
			return true, errors.Wrap(err, "failed to set txhashesByHeight map")
		}

		// remove old key
		if remove {
			err = s.oldTxhashesByHeightMap.Remove(ctx, key)
			if err != nil {
				return true, errors.Wrap(err, "failed to remove old txhashesByHeight map")
			}
		}
		s.Logger(ctx).Info("patching txhashesByHeight map", "key", key, "value", value)
		i++
		if i%1000 == 0 {
			s.Logger(ctx).Info("patching txhashesByHeight map", "count", i, "step", "migrate old store")
		}
		return false, nil
	})
	s.Logger(ctx).Info("patching txhashesByHeight map", "count", i, "step", "migrate old store")
	if err != nil {
		return errors.Wrap(err, "failed to walk through old txhashesByHeight map")
	}

	return nil
}

/*
func (s *EvmTxSubmodule) patchSequenceByHeightMap(ctx context.Context, oldSeq uint64) (err error) {

    i := 0
    err = s.oldSequenceByHeightMap.Walk(ctx, nil, func(key int64, value uint64) (stop bool, err error) {
        err = s.sequenceByHeightMap.Set(ctx, key, value)
        if err != nil {
            return true, errors.Wrap(err, "failed to set sequenceByHeight map")
        }

        // remove old key
        err = s.oldSequenceByHeightMap.Remove(ctx, key)
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


    i := 0
    err = s.oldAccountSequenceByHeightMap.Walk(ctx, nil, func(key collections.Triple[int64, sdk.AccAddress, uint64], value bool) (stop bool, err error) {
        err = s.accountSequenceByHeightMap.Set(ctx, key, value)
        if err != nil {
            return true, errors.Wrap(err, "failed to set accountSequenceByHeight map")
        }

        // remove old key
        err = s.oldAccountSequenceByHeightMap.Remove(ctx, key)
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
*/
