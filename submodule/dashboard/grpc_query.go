package dashboard

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"cosmossdk.io/collections"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/initia-labs/kvindexer/module/keeper"
	"github.com/initia-labs/kvindexer/submodule/dashboard/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = (*Querier)(nil)

type Querier struct {
	*keeper.Keeper
}

// NewQuerier return new Querier instance
func NewQuerier(k *keeper.Keeper) Querier {
	return Querier{k}
}

// NewAccounts implements types.QueryServer.
func (q Querier) NewAccounts(ctx context.Context, req *types.NewAccountsRequest) (*types.NewAccountsResponse, error) {
	if !enabled {
		return nil, status.Error(codes.Unavailable, fmt.Sprintf("cannot query: %s is disabled", submoduleName))
	}

	if req.Pagination != nil && limit > 0 {
		if req.Pagination.Limit > limit || req.Pagination.Limit == 0 {
			req.Pagination.Limit = limit
		}
	}

	accounts, pageRes, err := query.CollectionPaginate(
		ctx, accountMapByHeight,
		req.Pagination,
		func(key int64, value string) (aph *types.AccountsPerHeight, err error) {
			return &types.AccountsPerHeight{Height: key, Accounts: strings.Split(value, ",")}, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.NewAccountsResponse{
		Accounts:   accounts,
		Pagination: pageRes,
	}, err
}

func (q Querier) ChartData(ctx context.Context, req *types.ChartDataRequest) (*types.ChartDataResponse, error) {
	if !enabled {
		return nil, status.Error(codes.Unavailable, fmt.Sprintf("cannot query: %s is disabled", submoduleName))
	}

	valids := []int{7, 30, 90}
	limit := int(req.Limit)
	if !slices.Contains(valids, limit) {
		return nil, status.Error(codes.InvalidArgument, "limit is invalid, limit must be one of [7, 30, 90]")
	}

	now := timeToDateString(time.Now())
	start := timeToDateString(time.Now().AddDate(0, 0, -limit))

	rng := new(collections.Range[string]).StartInclusive(start).EndInclusive(now).Descending()

	// get tx count
	iter, err := txCountByDate.Iterate(ctx, rng)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer iter.Close()

	var txs []collections.KeyValue[string, uint64]
	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		txs = append(txs, collections.KeyValue[string, uint64]{
			Key:   kv.Key,
			Value: kv.Value,
		})
	}

	// get supply
	supplyIter, err := supplyByDate.Iterate(ctx, rng)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer supplyIter.Close()

	var supplies []collections.KeyValue[string, int64]
	for ; supplyIter.Valid(); supplyIter.Next() {
		kv, err := supplyIter.KeyValue()
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		supplies = append(supplies, collections.KeyValue[string, int64]{
			Key:   kv.Key,
			Value: int64(kv.Value),
		})
	}

	// get new accounts
	newAccountIter, err := newAccountCountMapByDate.Iterate(ctx, rng)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer newAccountIter.Close()

	var newAccounts []collections.KeyValue[string, int64]
	for ; newAccountIter.Valid(); newAccountIter.Next() {
		kv, err := newAccountIter.KeyValue()
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		newAccounts = append(newAccounts, collections.KeyValue[string, int64]{
			Key:   kv.Key,
			Value: int64(kv.Value),
		})
	}

	// get cumulative number of accounts
	accountIter, err := totalAccountBaseCountByDate.Iterate(ctx, rng)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer accountIter.Close()

	var accounts []collections.KeyValue[string, int64]
	for ; accountIter.Valid(); accountIter.Next() {
		kv, err := accountIter.KeyValue()
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		accounts = append(accounts, collections.KeyValue[string, int64]{
			Key:   kv.Key,
			Value: int64(kv.Value),
		})
	}

	// make a return data.
	dataMap := make(map[string]types.ChartData)

	for _, tx := range txs {
		dataMap[tx.Key] = types.ChartData{
			Date:    tx.Key,
			TxCount: int64(tx.Value),
		}
	}

	for _, supply := range supplies {
		if entry, ok := dataMap[supply.Key]; ok {
			entry.TotalValueLocked = supply.Value
			dataMap[supply.Key] = entry
		} else {
			dataMap[supply.Key] = types.ChartData{
				Date:             supply.Key,
				TotalValueLocked: supply.Value,
			}
		}
	}

	for _, newAccount := range newAccounts {
		if entry, ok := dataMap[newAccount.Key]; ok {
			entry.NewAccounts = newAccount.Value
			dataMap[newAccount.Key] = entry
		} else {
			dataMap[newAccount.Key] = types.ChartData{
				Date:        newAccount.Key,
				NewAccounts: newAccount.Value,
			}
		}
	}

	for _, account := range accounts {
		if entry, ok := dataMap[account.Key]; ok {
			entry.ActiveAccounts = account.Value
			dataMap[account.Key] = entry
		} else {
			dataMap[account.Key] = types.ChartData{
				Date:           account.Key,
				ActiveAccounts: account.Value,
			}
		}
	}
	
	var data []*types.ChartData
	for _, item := range dataMap {
		currentItem := item
		data = append(data, &currentItem)
	}

	return &types.ChartDataResponse{Data: data}, nil
}
