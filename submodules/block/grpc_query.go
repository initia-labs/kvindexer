package block

import (
	"context"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/initia-labs/kvindexer/submodules/block/types"
	"github.com/initia-labs/kvindexer/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	avg_denominator int64 = 1000
)

type Querier struct {
	BlockSubmodule
}

func NewQuerier(bs BlockSubmodule) Querier {
	return Querier{bs}
}

func (q Querier) Block(ctx context.Context, req *types.BlockRequest) (*types.BlockResponse, error) {
	block, err := q.blockByHeight.Get(ctx, req.Height)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &types.BlockResponse{
		Block: &block,
	}, nil

}

func (q Querier) Blocks(ctx context.Context, req *types.BlocksRequest) (*types.BlocksResponse, error) {
	util.ValidatePageRequest(req.Pagination)
	results, pageRes, err := query.CollectionPaginate(
		ctx,
		q.blockByHeight,
		req.Pagination,
		func(key int64, v types.Block) (*types.Block, error) {
			return &v, nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &types.BlocksResponse{
		Blocks:     results,
		Pagination: pageRes,
	}, nil
}

func (q Querier) AvgBlockTime(ctx context.Context, req *types.AvgBlockTimeRequest) (*types.AvgBlockTimeResponse, error) {
	rng := new(collections.Range[int64]).Descending()
	iter, err := q.blockByHeight.Iterate(ctx, rng)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer iter.Close()

	lastKV, err := iter.KeyValue()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	lastBlock := lastKV.Value

	base := lastBlock.Height - avg_denominator
	if base < 0 {
		base = 1 // from genesis
	}

	firstBlock, err := q.blockByHeight.Get(ctx, base)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	timeDiff := lastBlock.Timestamp.Sub(firstBlock.Timestamp).Milliseconds()
	heightDiff := lastBlock.Height - firstBlock.Height

	result := float64(timeDiff) / float64(heightDiff) / 1000

	return &types.AvgBlockTimeResponse{
		AvgBlockTime: result,
	}, nil
}
