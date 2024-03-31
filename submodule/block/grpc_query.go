package block

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/initia-labs/kvindexer/module/keeper"
	"github.com/initia-labs/kvindexer/submodule/block/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Querier struct {
	*keeper.Keeper
}

func NewQuerier(k *keeper.Keeper) Querier {
	return Querier{k}
}

func (q Querier) Blocks(ctx context.Context, req *types.BlocksRequest) (*types.BlocksResponse, error) {
	if !enabled {
		return nil, status.Error(codes.Unavailable, fmt.Sprintf("cannot query: %s is disabled", submoduleName))
	}

	results, pageRes, err := query.CollectionPaginate(
		ctx,
		blockByHeight,
		req.Pagination,
		func(key uint64, value []byte) (*types.Block, error) {
			block, err := makeBlock(value)
			if err != nil {
				return nil, err
			}
			return &block, nil
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
	if !enabled {
		return nil, status.Error(codes.Unavailable, fmt.Sprintf("cannot query: %s is disabled", submoduleName))
	}
	rng := new(collections.Range[uint64]).Descending()
	iter, err := blockByHeight.Iterate(ctx, rng)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer iter.Close()

	kv, err := iter.KeyValue()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var lastBlock types.Block
	err = json.Unmarshal(kv.Value, &lastBlock)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	fb, err := blockByHeight.Get(ctx, uint64(lastBlock.Height-1000)) // error when lastBlock.Height < 1000?
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var firstBlock types.Block
	err = json.Unmarshal(fb, &firstBlock)
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
