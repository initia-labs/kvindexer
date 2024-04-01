package tx

import (
	"context"
	"fmt"
	"strings"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/initia-labs/kvindexer/module/keeper"
	"github.com/initia-labs/kvindexer/submodule/tx/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = (*Querier)(nil)

type Querier struct {
	*keeper.Keeper
}

// Tx implements types.QueryServer.
func (q Querier) Tx(ctx context.Context, req *types.QueryTxRequest) (*types.QueryTxResponse, error) {
	if !enabled {
		return nil, status.Error(codes.Unavailable, fmt.Sprintf("cannot query: %s is disabled", submoduleName))
	}

	if req.TxHash == "" {
		return nil, status.Error(codes.InvalidArgument, "empty tx hash")
	}
	txHash := strings.ToLower(req.TxHash)

	tx, err := txMap.Get(ctx, txHash)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryTxResponse{Tx: &tx}, nil
}

// TxsByAccount implements types.QueryServer.
func (q Querier) TxsByAccount(ctx context.Context, req *types.QueryTxsByAccountRequest) (*types.QueryTxsResponse, error) {
	if !enabled {
		return nil, status.Error(codes.Unavailable, fmt.Sprintf("cannot query: %s is disabled", submoduleName))
	}

	if req.Pagination != nil && limit > 0 {
		if req.Pagination.Limit > limit || req.Pagination.Limit == 0 {
			req.Pagination.Limit = limit
		}
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "empty address")
	}
	acc, err := accAddressFromString(req.Address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	txHashes, pageRes, err := query.CollectionPaginate(ctx, txhashesByAccountMap, req.Pagination,
		func(_ collections.Pair[sdk.AccAddress, uint64], value string) (*string, error) {
			return &value, nil
		},
		query.WithCollectionPaginationPairPrefix[sdk.AccAddress, uint64](acc),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if len(txHashes) == 0 {
		return nil, status.Error(codes.NotFound, "no txs found")
	}

	txs := []*sdk.TxResponse{}
	for _, txHash := range txHashes {
		tx, err := txMap.Get(ctx, *txHash)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		txs = append(txs, &tx)
	}

	return &types.QueryTxsResponse{
		Txs:        txs,
		Pagination: pageRes,
	}, nil
}

// Txs implements types.QueryServer.
func (q Querier) Txs(ctx context.Context, req *types.QueryTxsRequest) (*types.QueryTxsResponse, error) {
	if !enabled {
		return nil, status.Error(codes.Unavailable, fmt.Sprintf("cannot query: %s is disabled", submoduleName))
	}

	if req.Pagination != nil && limit > 0 {
		if req.Pagination.Limit > limit || req.Pagination.Limit == 0 {
			req.Pagination.Limit = limit
		}
	}

	txHashes, pageRes, err := query.CollectionPaginate(ctx, txhashesBySequence, req.Pagination,
		func(_ uint64, value string) (*string, error) {
			return &value, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if len(txHashes) == 0 {
		return nil, status.Error(codes.NotFound, "no txs found")
	}

	txs := []*sdk.TxResponse{}
	for _, txHash := range txHashes {
		tx, err := txMap.Get(ctx, *txHash)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		txs = append(txs, &tx)
	}

	return &types.QueryTxsResponse{
		Txs:        txs,
		Pagination: pageRes,
	}, nil
}

// TxsByHeight implements types.QueryServer.
func (q Querier) TxsByHeight(ctx context.Context, req *types.QueryTxsByHeightRequest) (*types.QueryTxsResponse, error) {
	if !enabled {
		return nil, status.Error(codes.Unavailable, fmt.Sprintf("cannot query: %s is disabled", submoduleName))
	}

	if req.Pagination != nil && limit > 0 {
		if req.Pagination.Limit > limit || req.Pagination.Limit == 0 {
			req.Pagination.Limit = limit
		}
	}

	txHashes, pageRes, err := query.CollectionPaginate(ctx, txhashesByHeight, req.Pagination,
		func(_ collections.Pair[int64, uint64], value string) (*string, error) {
			return &value, nil
		},
		query.WithCollectionPaginationPairPrefix[int64, uint64](req.Height),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if len(txHashes) == 0 {
		return nil, status.Error(codes.NotFound, "no txs found")
	}

	txs := []*sdk.TxResponse{}
	for _, txHash := range txHashes {
		tx, err := txMap.Get(ctx, *txHash)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		txs = append(txs, &tx)
	}

	return &types.QueryTxsResponse{
		Txs:        txs,
		Pagination: pageRes,
	}, nil
}

// NewQuerier return new Querier instance
func NewQuerier(k *keeper.Keeper) Querier {
	return Querier{k}
}
