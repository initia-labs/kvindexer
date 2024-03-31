package dashboard

import (
	"context"
	"cosmossdk.io/collections"
	"cosmossdk.io/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/indexer/v2/config"
	"github.com/initia-labs/indexer/v2/module/keeper"
	"github.com/spf13/cast"
)

func processSupply(k *keeper.Keeper, ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock, cfg config.SubmoduleConfig) error {
	opBridgeId := cast.ToUint64(cfg["op-bridge-id"])
	l1Denom := cast.ToString(cfg["l1-denom"])

	date := timeToDateString(timestamp)

	_, err := supplyByDate.Get(ctx, date)
	if err != nil && !errors.IsOf(err, collections.ErrNotFound) {
		return errors.Wrap(err, "failed to get supply by date")
	}

	supplyMap := make(map[string]uint64)
	k.BankKeeper.IterateTotalSupply(ctx, func(coin sdk.Coin) bool {
		supplyMap[coin.Denom] = coin.Amount.Uint64()
		return false
	})
	denom := getOpDenom(opBridgeId, l1Denom)
	if err = supplyByDate.Set(ctx, date, supplyMap[denom]); err != nil {
		return errors.Wrap(err, "failed to set supply by date")
	}
	return nil
}
