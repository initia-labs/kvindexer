package pair

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/initia-labs/kvindexer/module/keeper"
	"github.com/initia-labs/kvindexer/submodule/pair/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = (*Querier)(nil)

type Querier struct {
	*keeper.Keeper
}

// Pairs implements types.QueryServer.
func (q Querier) Pairs(ctx context.Context, req *types.QueryPairsRequest) (*types.QueryPairsResponse, error) {
	if !enabled {
		return nil, status.Error(codes.Unavailable, fmt.Sprintf("cannot query: %s is disabled", submoduleName))
	}

	if req.Pagination != nil && limit > 0 {
		if req.Pagination.Limit > limit || req.Pagination.Limit == 0 {
			req.Pagination.Limit = limit
		}
	}

	pairs := []*types.Pair{}
	var targetMap *collections.Map[string, string]
	if req.IsFungible {
		targetMap = fungiblepairsMap
	} else {
		targetMap = nonFungiblepairsMap
	}
	_, pageRes, err := query.CollectionPaginate(ctx, targetMap, req.Pagination,
		func(key string, value string) (*string, error) {
			pair := types.Pair{
				L1: value,
				L2: key,
			}
			pairs = append(pairs, &pair)
			return &value, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPairsResponse{
		Pairs:      pairs,
		Pagination: pageRes,
	}, nil
}

// NewQuerier return new Querier instance
func NewQuerier(k *keeper.Keeper) Querier {
	return Querier{k}
}