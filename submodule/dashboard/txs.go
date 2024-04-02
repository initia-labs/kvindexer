package dashboard

import (
	"context"

	"cosmossdk.io/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/initia-labs/kvindexer/config"
	"github.com/initia-labs/kvindexer/module/keeper"
)

func processTxs(k *keeper.Keeper, ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock, cfg config.SubmoduleConfig) error {
	curTxsCount := uint64(len(req.Txs))
	err := updateUint64MapByDate(ctx, txCountByDate, curTxsCount, true)
	if err != nil {
		return errors.Wrap(err, "failed to update tx count by date")
	}
	return nil
}
