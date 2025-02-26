package tx

import (
	"context"
	"strings"

	"cosmossdk.io/collections"
	txdecode "cosmossdk.io/x/tx/decode"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/initia-labs/kvindexer/submodules/tx/types"
)

var _ types.QueryServer = (*Querier)(nil)

type Querier struct {
	TxSubmodule
}

func NewQuerier(sb TxSubmodule) types.QueryServer {
	return Querier{sb}
}

// Tx implements types.QueryServer.
func (q Querier) Tx(ctx context.Context, req *types.QueryTxRequest) (*types.QueryTxResponse, error) {
	if req.TxHash == "" {
		return nil, status.Error(codes.InvalidArgument, "empty tx hash")
	}
	txHash := strings.ToLower(req.TxHash)

	tx, err := q.txMap.Get(ctx, txHash)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryTxResponse{Tx: &tx}, nil
}

// TxsByAccount implements types.QueryServer.
func (q Querier) TxsByAccount(ctx context.Context, req *types.QueryTxsByAccountRequest) (*types.QueryTxsResponse, error) {
	if req.Account == "" {
		return nil, status.Error(codes.InvalidArgument, "empty account")
	}

	acc, err := accAddressFromString(req.Account)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	txHashes, pageRes, err := query.CollectionPaginate(ctx, q.txhashesByAccountMap, req.Pagination,
		func(_ collections.Pair[sdk.AccAddress, uint64], value string) (*string, error) {
			return &value, nil
		},
		query.WithCollectionPaginationPairPrefix[sdk.AccAddress, uint64](acc),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	txs := []*sdk.TxResponse{}
	for _, txHash := range txHashes {
		tx, err := q.txMap.Get(ctx, *txHash)
		if err != nil {
			q.Logger(ctx).Info("failed to get tx", "tx_hash", *txHash, "error", err)
			e := txdecode.ErrTxDecode
			tx = sdk.TxResponse{
				TxHash:    *txHash,
				Codespace: e.Codespace(),
				Code:      e.ABCICode(),
				RawLog:    e.Wrap(err.Error()).Error(),
			}
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
	txHashes, pageRes, err := query.CollectionPaginate(ctx, q.txhashesBySequenceMap, req.Pagination,
		func(_ uint64, value string) (*string, error) {
			return &value, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	txs := []*sdk.TxResponse{}
	for _, txHash := range txHashes {
		tx, err := q.txMap.Get(ctx, *txHash)
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
	txHashes, pageRes, err := query.CollectionPaginate(ctx, q.txhashesByHeightMap, req.Pagination,
		func(_ collections.Pair[int64, uint64], value string) (*string, error) {
			return &value, nil
		},
		query.WithCollectionPaginationPairPrefix[int64, uint64](req.Height),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	txs := []*sdk.TxResponse{}
	for _, txHash := range txHashes {
		tx, err := q.txMap.Get(ctx, *txHash)
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
