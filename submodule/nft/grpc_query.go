package nft

import (
	"context"
	"slices"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/initia-labs/kvindexer/module/keeper"
	"github.com/initia-labs/kvindexer/submodule/nft/types"
	"github.com/initia-labs/kvindexer/submodule/pair"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = (*Querier)(nil)

type Querier struct {
	*keeper.Keeper
}

// Collection implements types.QueryServer.
func (q Querier) Collection(ctx context.Context, req *types.QueryCollectionRequest) (*types.QueryCollectionResponse, error) {
	collectionSdkAddr, err := sdk.AccAddressFromBech32(req.CollectionAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	collection, err := collectionMap.Get(ctx, collectionSdkAddr)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &types.QueryCollectionResponse{
		Collection: &collection,
	}, nil
}

// Collections implements types.QueryServer.
func (q Querier) CollectionsByAccount(ctx context.Context, req *types.QueryCollectionsByAccountRequest) (*types.QueryCollectionsResponse, error) {
	accountSdkAddr, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	accountAddrString := accountSdkAddr.String()

	collectionSdkAddrs := []sdk.AccAddress{}
	_, pageRes, err := query.CollectionPaginate(ctx, collectionOwnerMap, req.Pagination,
		func(k collections.Pair[sdk.AccAddress, sdk.AccAddress], v uint64) (uint64, error) {
			if k.K2().String() == accountAddrString {
				collectionSdkAddrs = append(collectionSdkAddrs, k.K2())
			}
			return v, nil
		},
	)

	collections := []*types.IndexedCollection{}
	for _, collectionSdkAddr := range collectionSdkAddrs {
		collection, err := collectionMap.Get(ctx, collectionSdkAddr)
		if err != nil {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		collections = append(collections, &collection)
	}

	return &types.QueryCollectionsResponse{
		Collections: collections,
		Pagination:  pageRes,
	}, nil
}

// TokensByCollection implements types.QueryServer.
func (q Querier) TokensByCollection(ctx context.Context, req *types.QueryTokensByCollectionRequest) (*types.QueryTokensResponse, error) {
	if req.TokenId == "" {
		return getTokensByCollection(q.Keeper, ctx, req)
	}
	return getTokensByCollectionAndTokenId(q.Keeper, ctx, req)
}

// TokensByAccount implements types.QueryServer.
func (q Querier) TokensByAccount(ctx context.Context, req *types.QueryTokensByAccountRequest) (*types.QueryTokensResponse, error) {
	if req.CollectionAddr == "" {
		return getTokensByAccount(q.Keeper, ctx, req)
	}
	if req.TokenId == "" {
		return getTokensByAccountAndCollection(q.Keeper, ctx, req)
	}
	return getTokensByAccountCollectionAndTokenId(q.Keeper, ctx, req)
}

// NewQuerier return new Querier instance
func NewQuerier(k *keeper.Keeper) Querier {
	return Querier{k}
}

func getCollectionNameFromPairSubmodule(ctx context.Context, collName string) (string, error) {
	return pair.GetPair(ctx, false, collName)
}

func getTokensByCollection(_ *keeper.Keeper, ctx context.Context, req *types.QueryTokensByCollectionRequest) (*types.QueryTokensResponse, error) {
	collAddr, err := sdk.AccAddressFromBech32(req.CollectionAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	res, pageRes, err := query.CollectionPaginate(ctx, tokenMap, req.Pagination,
		func(k collections.Pair[sdk.AccAddress, string], v types.IndexedToken) (*types.IndexedToken, error) {
			if slices.Equal(k.K1(), collAddr) {
				return &v, nil
			}
			return nil, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryTokensResponse{
		Tokens:     res,
		Pagination: pageRes,
	}, nil

}

func getTokensByCollectionAndTokenId(_ *keeper.Keeper, ctx context.Context, req *types.QueryTokensByCollectionRequest) (*types.QueryTokensResponse, error) {
	collAddr, err := sdk.AccAddressFromBech32(req.CollectionAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	token, err := tokenMap.Get(ctx, collections.Join(collAddr, req.TokenId))
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &types.QueryTokensResponse{
		Tokens: []*types.IndexedToken{&token},
	}, nil
}

func getTokensByAccount(k *keeper.Keeper, _ context.Context, req *types.QueryTokensByAccountRequest) (*types.QueryTokensResponse, error) {
	ownerAddr, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	ownerAddrStr := ownerAddr.String()

	store := k.GetStore()
	ownerStore := prefix.NewStore(runtime.KVStoreAdapter(store), prefixTokenOwnerIndex)

	res, pageRes, err := query.GenericFilteredPaginate(
		k.GetCodec(),   /*codec*/
		ownerStore,     /* store */
		req.Pagination, /* page request */
		func(key []byte, val *types.IndexedToken) (*types.IndexedToken, error) {
			if val.OwnerAddr != ownerAddrStr {
				return nil, nil
			}
			return val, nil
		}, /* onResult */
		func() *types.IndexedToken {
			return &types.IndexedToken{}
		}, /* constructor */
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryTokensResponse{
		Tokens:     res,
		Pagination: pageRes,
	}, nil
}

func getTokensByAccountAndCollection(_ *keeper.Keeper, ctx context.Context, req *types.QueryTokensByAccountRequest) (*types.QueryTokensResponse, error) {
	collAddr, err := sdk.AccAddressFromBech32(req.CollectionAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ownerAddr, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	ownerAddrStr := ownerAddr.String()

	res, pageRes, err := query.CollectionPaginate(ctx, tokenMap, req.Pagination,
		func(k collections.Pair[sdk.AccAddress, string], v types.IndexedToken) (*types.IndexedToken, error) {
			if slices.Equal(k.K1(), collAddr) && (v.OwnerAddr == ownerAddrStr) {
				return &v, nil
			}
			return nil, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryTokensResponse{
		Tokens:     res,
		Pagination: pageRes,
	}, nil
}

func getTokensByAccountCollectionAndTokenId(_ *keeper.Keeper, ctx context.Context, req *types.QueryTokensByAccountRequest) (*types.QueryTokensResponse, error) {
	collAddr, err := sdk.AccAddressFromBech32(req.CollectionAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	token, err := tokenMap.Get(ctx, collections.Join(collAddr, req.TokenId))
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	if token.OwnerAddr != req.Account {
		return nil, status.Error(codes.NotFound, "token not found")
	}

	return &types.QueryTokensResponse{
		Tokens: []*types.IndexedToken{&token},
	}, nil
}
