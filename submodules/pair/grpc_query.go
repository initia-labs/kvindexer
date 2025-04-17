package pair

import (
	"context"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/initia-labs/kvindexer/pair/types"
	"github.com/initia-labs/kvindexer/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = (*Querier)(nil)

type Querier struct {
	PairSubmodule
}

func NewQuerier(bs PairSubmodule) Querier {
	return Querier{bs}
}

// Pairs implements types.QueryServer.
func (q Querier) Pairs(ctx context.Context, req *types.QueryPairsRequest) (*types.QueryPairsResponse, error) {
	util.ValidatePageRequest(req.Pagination)
	pairs := []*types.Pair{}
	var targetMap *collections.Map[string, string]
	if req.IsNonFungible {
		targetMap = q.nonFungiblePairsMap
	} else {
		targetMap = q.fungiblePairsMap
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
