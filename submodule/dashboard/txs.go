package dashboard

import (
	"context"

	"cosmossdk.io/collections"
	"cosmossdk.io/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/initia-labs/indexer/v2/config"
	"github.com/initia-labs/indexer/v2/module/keeper"
)

func processTxs(k *keeper.Keeper, ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock, cfg config.SubmoduleConfig) error {
	// set cumulative number of txs.
	prev := timestamp.AddDate(0, 0, -1)
	prevCount, err := txTotalCountByDate.Get(ctx, timeToDateString(prev))
	if err != nil {
		prevCount = 0
	}

	date := timeToDateString(timestamp)

	curTxsCount := uint64(len(req.Txs))

	lastTxCount, err := txCountByDate.Get(ctx, date)

	if err != nil && !errors.IsOf(err, collections.ErrNotFound) {
		return errors.Wrap(err, "failed to get tx count by date")
	}
	if err = txCountByDate.Set(ctx, date, lastTxCount+curTxsCount); err != nil {
		return errors.Wrap(err, "failed to set tx count by date")
	}
	if err = txTotalCountByDate.Set(ctx, date, prevCount+lastTxCount+curTxsCount); err != nil {
		return errors.Wrap(err, "failed to set tx total count by date")
	}
	return nil
}
