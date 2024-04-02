package dashboard

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/errors"
	"github.com/initia-labs/kvindexer/submodule/dashboard/types"
	"golang.org/x/crypto/sha3"
)

const dateFmt = "2006-01-02"

var dateOldEnough = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func timeToDateString(t time.Time) string {
	return t.Format(dateFmt)
}

func getOpDenom(bridgeID uint64, l1Denom string) string {
	bridgeIDBuf := make([]byte, 8)
	for i := 0; i < 8; i++ {
		bridgeIDBuf[7-i] = byte(bridgeID >> (i * 8))
	}
	hash := sha3.Sum256(append(bridgeIDBuf, []byte(l1Denom)...))

	return fmt.Sprintf("l2/%s", hex.EncodeToString(hash[:]))
}

func getPastValue(getter types.Uint64Map, ctx context.Context, begin_before int) (prevCount uint64, err error) {
	for prevCount == 0 {
		prev := timestamp.AddDate(0, 0, -begin_before)
		if prev.Before(dateOldEnough) {
			return 0, nil
		}
		prevCount, err = getter.Get(ctx, timeToDateString(prev))
		if err != nil {
			if !errors.IsOf(err, collections.ErrNotFound) {
				return 0, err
			}
		} else {
			break
		}
		begin_before++
	}
	return prevCount, nil
}

func updateUint64MapByDate(ctx context.Context, umap types.Uint64Map, value uint64, isCumulativeByDate bool) error {
	date := timeToDateString(timestamp)
	prevCount, err := umap.Get(ctx, date)
	if err != nil {
		if !errors.IsOf(err, collections.ErrNotFound) {
			return errors.Wrap(err, "failed to get from map")
		}
		// pick previous count from the last value exising day since yesterday in revserse order
		if isCumulativeByDate {
			prevCount, err = getPastValue(umap, ctx, -1)
			if err != nil {
				return errors.Wrap(err, "failed to get past value from map")
			}
		} else {
			prevCount = 0
		}
		// if there is no today's count set it as previous to prevent missing today's count
		err = umap.Set(ctx, date, prevCount)
		if err != nil {
			return errors.Wrap(err, "failed to set map")
		}
	}

	// no need to update if current tx count is 0
	if value == 0 {
		return nil
	}
	if err = umap.Set(ctx, date, prevCount+value); err != nil {
		return errors.Wrap(err, "failed to update map")
	}
	return nil
}
