package dashboard

import (
	"bytes"
	"context"
	"slices"
	"strings"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
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

	err := updateTotalBaseAccountByDate(k, ctx)
	if err != nil {
		return errors.Wrap(err, "failed to update total base account by date")
	}
	return nil
}

func updateTotalBaseAccountByDate(k *keeper.Keeper, ctx context.Context) error {
	// set cumulative number of accounts
	lastAccNum, err := getLastAccountNumber(ctx)
	if err != nil {
		return err
	}

	date := timeToDateString(timestamp)
	count, err := totalAccountBaseCountByDate.Get(ctx, date)
	if err != nil {
		if errors.IsOf(err, collections.ErrNotFound) {
			prev := timestamp.AddDate(0, 0, -1)
			prevCount, err := totalAccountBaseCountByDate.Get(ctx, timeToDateString(prev))
			if err != nil {
				prevCount = 0 // fail to get previous total number of accounts by date
			}
			count = prevCount
		} else {
			return errors.Wrap(err, "failed to get total accounts by date")
		}
	}

	currentAccNum, _ := k.AccountKeeper.AccountNumber.Peek(ctx)
	rng := new(collections.Range[uint64]).StartInclusive(lastAccNum).EndExclusive(currentAccNum)
	iter, err := k.AccountKeeper.Accounts.Indexes.Number.Iterate(ctx, rng)
	if err != nil {
		return err
	}
	defer iter.Close()
	_ = indexes.ScanValues(ctx, k.AccountKeeper.Accounts, iter, func(acc sdk.AccountI) bool {
		if _, ok := acc.(*authtypes.BaseAccount); ok {
			count += 1
		}
		return false
	})

	if err = totalAccountBaseCountByDate.Set(ctx, date, uint64(count)); err != nil {
		return errors.Wrap(err, "failed to set total accounts by date")
	}

	err = lastAccountNumber.Set(ctx, currentAccNum)
	if err != nil {
		return errors.Wrap(err, "failed to set last account number")
	}
	return nil
}

func getLastAccountNumber(ctx context.Context) (uint64, error) {
	num, err := lastAccountNumber.Get(ctx)
	if err != nil {
		if errors.IsOf(err, collections.ErrNotFound) {
			return 0, nil
		}
		return 0, errors.Wrap(err, "failed to get last account number")
	}
	return num, nil
}
