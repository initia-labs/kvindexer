package dashboard

import (
	"bytes"
	"context"
	"slices"
	"strings"

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

	if len(newAccs) > 0 {
		// insert new accouts by height
		if err := accountMapByHeight.Set(ctx, height, strings.Join(newAccs, ",")); err != nil {
			return errors.Wrap(err, "failed to set new accounts by height")
		}
	}
	err := updateUint64MapByDate(ctx, newAccountCountMapByDate, uint64(len(newAccs)), false)
	if err != nil {
		return errors.Wrap(err, "failed to update new base account by date")
	}

	err = updateUint64MapByDate(ctx, totalAccountBaseCountByDate, uint64(len(newAccs)), true)
	if err != nil {
		return errors.Wrap(err, "failed to update total base account by date")
	}
	return nil
}
