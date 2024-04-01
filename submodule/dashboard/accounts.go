package dashboard

import (
	"bytes"
	"context"
	"slices"
	"strings"

	"cosmossdk.io/collections"
	"cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/initia-labs/kvindexer/config"
	"github.com/initia-labs/kvindexer/module/keeper"
)

func processAccounts(k *keeper.Keeper, ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair, cfg config.SubmoduleConfig) error {
	newAccs := []string{}

	// gather all new accounts from changeSet
	for _, kv := range changeSet {
		if kv.StoreKey != string(authtypes.StoreKey) || kv.Delete {
			continue
		}
		if !bytes.HasPrefix(kv.Key, authtypes.AddressStoreKeyPrefix) {
			continue
		}

		var any codectypes.Any
		err := any.Unmarshal(kv.Value)
		if err != nil {
			k.Logger(ctx).Error("failed to unmarshal to Any", "err", err)
			k.Logger(ctx).Debug("kvpair", "key", kv.Key, "value", kv.Value)
			return err
		}

		// if the account is not BaseAccount, skip
		if any.TypeUrl != "/cosmos.auth.v1beta1.BaseAccount" {
			continue
		}

		var acc sdk.AccountI = &authtypes.BaseAccount{}
		err = k.GetCodec().Unmarshal(any.Value, acc)
		if err != nil {
			k.Logger(ctx).Error("failed to unmarshal to BaseAccount", "err", err)
			k.Logger(ctx).Debug("any", "key", any.TypeUrl, "value", any.Value)
		}

		if acc.GetPubKey() != nil && acc.GetSequence() == 1 {
			if !slices.Contains(newAccs, acc.GetAddress().String()) {
				newAccs = append(newAccs, acc.GetAddress().String())
			}
		}
	}
	k.Logger(ctx).Debug(submoduleName, "height", height, "new-accounts", strings.Join(newAccs, ","))

	date := timeToDateString(timestamp)
	if len(newAccs) > 0 {
		// update new accounts by date
		num, err := newAccountCountMapByDate.Get(ctx, date)
		if err != nil && !errors.IsOf(err, collections.ErrNotFound) {
			return errors.Wrap(err, "failed to get new accounts by date")
		}
		if err = newAccountCountMapByDate.Set(ctx, date, num+uint64(len(newAccs))); err != nil {
			return errors.Wrap(err, "failed to set new accounts by date")
		}
		// insert new accouts by height
		if err = accountMapByHeight.Set(ctx, height, strings.Join(newAccs, ",")); err != nil {
			return errors.Wrap(err, "failed to set new accounts by height")
		}
	}

	// set cumulative number of accounts
	//prev := timestamp.AddDate(0, 0, -1)
	//prevCount, err := totalAccountCountByDate.Get(ctx, timeToDateString(prev))
	//if err != nil {
	//	prevCount = 0 // fail to get previous total number of accounts by date
	//}
	//
	//n, _ := k.AccountKeeper.AccountNumber.Peek(ctx)
	//rng := new(collections.Range[uint64]).StartExclusive(prevCount).EndInclusive(n)
	//iter, err := k.AccountKeeper.Accounts.Indexes.Number.Iterate(ctx, rng)
	//if err != nil {
	//	return err
	//}
	//defer iter.Close()
	//
	//var count = 0
	//_ = indexes.ScanValues(ctx, k.AccountKeeper.Accounts, iter, func(acc sdk.AccountI) bool {
	//	if _, ok := acc.(*authtypes.BaseAccount); ok {
	//		count += 1
	//	}
	//	return false
	//})
	//
	//if err = totalAccountCountByDate.Set(ctx, date, uint64(count)); err != nil {
	//	return errors.Wrap(err, "failed to set total accounts by date")
	//}

	var count = 0
	k.AccountKeeper.IterateAccounts(ctx, func(account sdk.AccountI) bool {
		if _, ok := account.(*authtypes.BaseAccount); ok {
			count++
		}
		return false
	})
	if err := totalAccountCountByDate.Set(ctx, date, uint64(count)); err != nil {
		return errors.Wrap(err, "failed to set total accounts by date")
	}

	return nil
}
