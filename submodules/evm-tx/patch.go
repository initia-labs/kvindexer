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

	/** unncecessary
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
	s.Logger(ctx).Info("start patching sequence")

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
	// flush
	if err = s.keeper.Write(); err != nil {
		return 0, errors.Wrap(err, "failed to write current sequence value")
	}

	return oldval, nil
}

func (s *EvmTxSubmodule) patchTxMap(ctx context.Context) (err error) {
	s.Logger(ctx).Info("start patching txMap")

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
				s.Logger(ctx).Info("[DEBUG] FIRST ITEM", "key", key)
			}

			// remove old key
			err = s.oldTxMap.Remove(ctx, key)
			if err != nil {
				return res, errors.Wrap(err, "failed to remove old tx map")
			}

			return value, nil
		})
		if err != nil {
			return errors.Wrap(err, "failed to walk through old tx map")
		}
		i += len(txMap)

		s.Logger(ctx).Info("patching tx map", "count", len(txMap), "nextKey", string(pageRes.NextKey))

		for k, v := range txMap {
			err = s.txMap.Set(ctx, k, v)
			if err != nil {
				return errors.Wrap(err, "failed to set tx map")
			}
		}
		s.Logger(ctx).Info("migrate tx map", "count", len(txMap))
		clear(txMap)

		// flush
		if err = s.keeper.Write(); err != nil {
			return errors.Wrap(err, "failed to write tx map")
		}
		s.Logger(ctx).Info("write tx map", "count", len(txMap))

		//pageReq.Key = pageRes.NextKey
		if pageRes == nil || len(pageRes.NextKey) == 0 {
			break
		}
	}
	s.Logger(ctx).Info("patching tx map done", "count", i)
	if err != nil {
		return errors.Wrap(err, "failed to walk through old tx map")
	}

	return nil
}

func (s *EvmTxSubmodule) patcAccountSequenceMap(ctx context.Context) (err error) {
	s.Logger(ctx).Info("start patching accountSequenceMap")

	i := 0
	pageReq := &query.PageRequest{
		Key:        nil,
		Offset:     0,
		Limit:      iterUnit,
		CountTotal: false,
	}

	for {
		accseqMap := make(map[string]uint64)
		isFirst := true
		_, pageRes, err := query.CollectionPaginate(ctx, s.oldAccountSequenceMap, pageReq, func(key sdk.AccAddress, value uint64) (res uint64, err error) {
			accseqMap[key.String()] = value

			if isFirst {
				isFirst = false
				s.Logger(ctx).Info("[DEBUG] FIRST ITEM", "key", key.String())
			}

			// remove old key
			err = s.oldAccountSequenceMap.Remove(ctx, key)
			if err != nil {
				return res, errors.Wrap(err, "failed to remove old accountSequence map")
			}
			return value, nil
		})
		if err != nil {
			return errors.Wrap(err, "failed to walk through old accountSequence map")
		}
		i += len(accseqMap)
		s.Logger(ctx).Info("patching accountSequence map", "count", len(accseqMap), "nextKey", string(pageRes.NextKey))

		for straddr, oldval := range accseqMap {
			addr, err := sdk.AccAddressFromBech32(straddr)
			if err != nil {
				s.Logger(ctx).Error("failed to convert address", "err", err, "str", straddr)
				return errors.Wrap(err, "failed to convert address")
			}
			var curval uint64
			var has bool
			has, err = s.accountSequenceMap.Has(ctx, addr)
			if err != nil {
				s.Logger(ctx).Error("failed to check accountSequence map", "err", err, "addr", straddr)
				return errors.Wrap(err, "failed to check accountSequence map")
			}
			if has {
				curval, err = s.accountSequenceMap.Get(ctx, addr)
				if err != nil {
					s.Logger(ctx).Error("failed to get current accountSequence map", "err", err, "addr", straddr)
					return errors.Wrap(err, "failed to get current accountSequence map")
				}
			} else {
				curval = 0
			}

			err = s.accountSequenceMap.Set(ctx, addr, curval+oldval)
			if err != nil {
				return errors.Wrap(err, "failed to set accountSequence map")
			}
		}
		s.Logger(ctx).Info("migrate accountSequence map", "count", len(accseqMap))
		clear(accseqMap)

		// flush
		if err = s.keeper.Write(); err != nil {
			return errors.Wrap(err, "failed to write accountSequence map")
		}
		s.Logger(ctx).Info("write accountSequence map", "count", len(accseqMap))

		//pageReq.Key = pageRes.NextKey
		if pageRes == nil || len(pageRes.NextKey) == 0 {
			break
		}
	}
	s.Logger(ctx).Info("patching account sequence map", "count", i)
	if err != nil {
		return errors.Wrap(err, "failed to walk through old accountSequence map")
	}

	return nil
}

func (s *EvmTxSubmodule) patchTxhashesByAccountMap(ctx context.Context) (err error) {
	s.Logger(ctx).Info("start patching txhashesByAccountMap")

	// just extracted map from current store
	curMap := make(map[collections.Pair[sdk.AccAddress, uint64]]string)
	// key: address, value: last seq from old store
	latestSeq := make(map[string]uint64)

	pageReq := &query.PageRequest{
		Key:        nil,
		Offset:     0,
		Limit:      iterUnit,
		CountTotal: false,
	}

	// pop all current txhashesByAccountMap
	for {
		isFirst := true
		_, pageRes, err := query.CollectionPaginate(ctx, s.txhashesByAccountMap, pageReq, func(key collections.Pair[sdk.AccAddress, uint64], value string) (res string, err error) {
			curMap[key] = value

			if isFirst {
				isFirst = false
				s.Logger(ctx).Info("[DEBUG] FIRST ITEM(CUR)", "k1", key.K1().String(), "k2", key.K2(), "v", value)
			}

			// remove old key
			err = s.txhashesByAccountMap.Remove(ctx, key)
			if err != nil {
				return res, errors.Wrap(err, "failed to remove current txhashedByAccount map")
			}
			return res, nil
		})
		if err != nil {
			return errors.Wrap(err, "failed to walk through current txhashedByAccount map")
		}

		// flush
		if err = s.keeper.Write(); err != nil {
			return errors.Wrap(err, "failed to write current txhashedByAccount map")
		}
		s.Logger(ctx).Info("patching txhashedByAccount map", "count", len(curMap), "nextKey", string(pageRes.NextKey))

		if pageRes == nil || len(pageRes.NextKey) == 0 {
			break
		}
	}
	s.Logger(ctx).Info("pop from current txhashedByAccount map", "mapcount", len(curMap))

	i := 0
	for {
		isFirst := true
		// migrate old txhashesByAccountMap to current one
		_, pageRes, err := query.CollectionPaginate(ctx, s.oldTxhashesByAccountMap, pageReq, func(key collections.Pair[sdk.AccAddress, uint64], value string) (res string, err error) {
			if isFirst {
				isFirst = false
				s.Logger(ctx).Info("[DEBUG] FIRST ITEM(OLD)", "k1", key.K1().String(), "k2", key.K2(), "v", value)
			}

			// remove old key
			err = s.oldTxhashesByAccountMap.Remove(ctx, key)
			if err != nil {
				return res, errors.Wrap(err, "failed to remove old txhashedByAccount map")
			}

			k1 := key.K1().String() // k1 := address from key
			lseq, ok := latestSeq[k1]
			if !ok {
				latestSeq[k1] = key.K2()
			} else {
				// if the old seq is greater than the current one, then we need to update the current one
				if key.K2() > lseq {
					latestSeq[k1] = key.K2()
				}
			}

			err = s.txhashesByAccountMap.Set(ctx, key, value)
			if err != nil {
				return res, errors.Wrap(err, "failed to set txhashedByAccount map")
			}
			i++
			return value, nil
		})
		if err != nil {
			return errors.Wrap(err, "failed to walk through old txhashedByAccount map")
		}

		// flush
		if err = s.keeper.Write(); err != nil {
			return errors.Wrap(err, "failed to write old txhashedByAccount map")
		}
		s.Logger(ctx).Info("write old txhashedByAccount map", "count", i)
		if pageRes == nil || len(pageRes.NextKey) == 0 {
			break
		}
	}

	// re-insert current txhashesByAccountMap
	i = 0
	for k, v := range curMap {
		addr := k.K1()
		lastSeq := latestSeq[addr.String()]
		if lastSeq == 0 {
			// not found - it means this account is came up after the upgrade
			err = s.txhashesByAccountMap.Set(ctx, k, v)
		} else {
			// found - it means this account is already in the old store
			// we need to set the last seq + current seq
			err = s.txhashesByAccountMap.Set(ctx, collections.Join(addr, lastSeq+k.K2()), v)
		}
		if err != nil {
			return errors.Wrap(err, "failed to set txhashedByAccount map")
		}
		i++
		if i%iterUnit == 0 {
			s.Logger(ctx).Info("patching txhashedByAccount map", "count", i, "step", "re-insert cur store")
			// flush
			if err = s.keeper.Write(); err != nil {
				s.Logger(ctx).Error("failed to write txhashedByAccount map", "err", err, "count", i, "step", "re-insert cur store")

				return errors.Wrap(err, "failed to write txhashedByAccount map")
			}
		}
	}
	// in case of there are still some left to be written
	if err = s.keeper.Write(); err != nil {
		s.Logger(ctx).Error("failed to write txhashedByAccount map", "err", err, "count", i, "step", "re-insert cur store")
		return errors.Wrap(err, "failed to write txhashedByAccount map")
	}

	s.Logger(ctx).Info("patched txhashedByAccount map", "count", i, "step", "re-insert cur store")

	return nil
}

func (s *EvmTxSubmodule) patchTxhashesBySequenceMap(ctx context.Context, oldSeq uint64) (err error) {
	s.Logger(ctx).Info("start patching txhashesBySequenceMap")

	i := 0
	pageReq := &query.PageRequest{
		Key:        nil,
		Offset:     0,
		Limit:      iterUnit,
		CountTotal: false,
	}

	seqMap := make(map[uint64]string)
	for {
		_, pageRes, err := query.CollectionPaginate(ctx, s.txhashesBySequenceMap, pageReq, func(key uint64, value string) (res string, err error) {
			seqMap[key] = value

			// remove old key
			err = s.txhashesBySequenceMap.Remove(ctx, key)
			if err != nil {
				return res, errors.Wrap(err, "failed to remove old txhashesBySequence map")
			}

			return value, nil
		})
		if err != nil {
			return errors.Wrap(err, "failed to walk through old txhashesBySequence map")
		}
		i += len(seqMap)
		s.Logger(ctx).Info("patching txhashesBySequence map", "count", len(seqMap), "nextKey", string(pageRes.NextKey))

		// flush
		if err = s.keeper.Write(); err != nil {
			return errors.Wrap(err, "failed to write old txhashesBySequence map")
		}
		if pageRes == nil || len(pageRes.NextKey) == 0 {
			break
		}
	}
	s.Logger(ctx).Info("pop txhashesBySequence map", "count", i)

	i = 0
	for k, v := range seqMap {
		err = s.txhashesBySequenceMap.Set(ctx, k+oldSeq, v)
		if err != nil {
			return errors.Wrap(err, "failed to set txhashesBySequence map")
		}
		s.Logger(ctx).Info("patching txhashesBySequence map", "key", k, "value", v)
		i++
		if i%iterUnit == 0 {
			s.Logger(ctx).Info("patching txhashesBySequence map", "count", i, "step", "re-insert cur store")
			// flush
			if err = s.keeper.Write(); err != nil {
				s.Logger(ctx).Error("failed to write txhashesBySequence map", "err", err, "count", i, "step", "re-insert cur store")
				return errors.Wrap(err, "failed to write txhashesBySequence map")
			}
		}
	}
	// in case of there are still some left to be written
	if err = s.keeper.Write(); err != nil {
		s.Logger(ctx).Error("failed to write txhashesBySequence map", "err", err, "count", i, "step", "re-insert cur store")
	}
	s.Logger(ctx).Info("pushback txhashesBySequence map", "count", i)

	i = 0
	for {
		_, pageRes, err := query.CollectionPaginate(ctx, s.oldTxhashesBySequenceMap, pageReq, func(key uint64, value string) (res string, err error) {
			err = s.txhashesBySequenceMap.Set(ctx, key, value)
			if err != nil {
				return res, errors.Wrap(err, "failed to set txhashesBySequence map")
			}

			// remove old key
			err = s.oldTxhashesBySequenceMap.Remove(ctx, key)
			if err != nil {
				return res, errors.Wrap(err, "failed to remove old txhashesBySequence map")
			}
			i++
			if i%iterUnit == 0 {
				s.Logger(ctx).Info("patching txhashesBySequence map", "count", i, "step", "migrate old store")
			}
			return value, nil
		})

		if err = s.keeper.Write(); err != nil {
			s.Logger(ctx).Error("failed to write txhashesBySequence map", "err", err, "count", i, "step", "migrate old store")
			return errors.Wrap(err, "failed to write txhashesBySequence map")
		}

		if err != nil {
			return errors.Wrap(err, "failed to walk through old txhashesBySequence map")
		}

		if pageRes == nil || len(pageRes.NextKey) == 0 {
			break
		}
	}
	// in case of there are still some left to be written
	if err = s.keeper.Write(); err != nil {
		s.Logger(ctx).Error("failed to write txhashesBySequence map", "err", err, "count", i, "step", "migrate old store")
		return errors.Wrap(err, "failed to write txhashesBySequence map")
	}
	s.Logger(ctx).Info("patched txhashesBySequence map", "migrate", i, "pushback", len(seqMap), "total", i+len(seqMap))

	return nil
}

func (s *EvmTxSubmodule) patchTxhashesByHeightMap(ctx context.Context) (err error) {
	s.Logger(ctx).Info("start patching txhashesByHeightMap")

	i := 0
	pageReq := &query.PageRequest{
		Key:        nil,
		Offset:     0,
		Limit:      iterUnit,
		CountTotal: false,
	}
	for {
		txhashesByHeightMap := make(map[collections.Pair[int64, uint64]]string)
		isFirst := true
		_, pageRes, err := query.CollectionPaginate(ctx, s.oldTxhashesByHeightMap, pageReq, func(key collections.Pair[int64, uint64], value string) (res string, err error) {
			txhashesByHeightMap[key] = value

			if isFirst {
				isFirst = false
				s.Logger(ctx).Info("[DEBUG] FIRST ITEM", "k1", key.K1(), "k2", key.K2(), "v", value)
			}

			// remove old key
			err = s.oldTxhashesByHeightMap.Remove(ctx, key)
			if err != nil {
				return res, errors.Wrap(err, "failed to remove current txhashesByHeight map")
			}

			return value, nil
		})
		if err != nil {
			return errors.Wrap(err, "failed to walk through current txhashesByHeight map")
		}
		i += len(txhashesByHeightMap)

		s.Logger(ctx).Info("patching txhashesByHeight map", "count", len(txhashesByHeightMap), "nextKey", string(pageRes.NextKey))

		for k, v := range txhashesByHeightMap {
			err = s.txhashesByHeightMap.Set(ctx, k, v)
			if err != nil {
				return errors.Wrap(err, "failed to set txhashesByHeight map")
			}
		}
		s.Logger(ctx).Info("migrate txhashesByHeight map", "count", len(txhashesByHeightMap))
		clear(txhashesByHeightMap)

		// flush
		if err = s.keeper.Write(); err != nil {
			return errors.Wrap(err, "failed to write txhashesByHeight map")
		}
		s.Logger(ctx).Info("write txhashesByHeight map", "count", len(txhashesByHeightMap))

		//pageReq.Key = pageRes.NextKey
		if pageRes == nil || len(pageRes.NextKey) == 0 {
			break
		}
	}

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
