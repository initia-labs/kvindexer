package nft

import (
	"context"
	"fmt"
	"slices"

	"cosmossdk.io/collections"
	cosmoserr "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
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

func handleCollectionErr(err error) error {
	if err == nil {
		return nil
	}
	if cosmoserr.IsOf(err, collections.ErrNotFound) {
		return status.Error(codes.NotFound, err.Error())
	}
	return status.Error(codes.Internal, err.Error())
}

// Collection implements types.QueryServer.
func (q Querier) Collection(ctx context.Context, req *types.QueryCollectionRequest) (*types.QueryCollectionResponse, error) {
	if !enabled {
		return nil, status.Error(codes.Unavailable, fmt.Sprintf("cannot query: %s is disabled", submoduleName))
	}
	q.Logger(ctx).Warn("[DEBUG] QueryCollection", "req", req)

	collectionAddr, err := getVMAddress(q.GetAddressCodec(), req.CollectionAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	collectionSdkAddr := getCosmosAddress(collectionAddr)

	collection, err := collectionMap.Get(ctx, collectionSdkAddr)
	if err != nil {
		return nil, handleCollectionErr(err)
	}

	return &types.QueryCollectionResponse{
		Collection: &collection,
	}, nil
}

// Collections implements types.QueryServer.
func (q Querier) Collections(ctx context.Context, req *types.QueryCollectionsRequest) (*types.QueryCollectionsResponse, error) {
	if !enabled {
		return nil, status.Error(codes.Unavailable, fmt.Sprintf("cannot query: %s is disabled", submoduleName))
	}

	if req.Pagination != nil && limit > 0 {
		if req.Pagination.Limit > limit || req.Pagination.Limit == 0 {
			req.Pagination.Limit = limit
		}
	}

	accountAddr, err := getVMAddress(q.GetAddressCodec(), req.Account)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	accountSdkAddr := getCosmosAddress(accountAddr)
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
			return nil, handleCollectionErr(err)
		}
		collections = append(collections, &collection)
	}

	return &types.QueryCollectionsResponse{
		Collections: collections,
		Pagination:  pageRes,
	}, nil
}

// Tokens implements types.QueryServer.
func (q Querier) Tokens(ctx context.Context, req *types.QueryTokensRequest) (*types.QueryTokensResponse, error) {
	if !enabled {
		return nil, status.Error(codes.Unavailable, fmt.Sprintf("cannot query: %s is disabled", submoduleName))
	}

	if req.Pagination != nil && limit > 0 {
		if req.Pagination.Limit > limit || req.Pagination.Limit == 0 {
			req.Pagination.Limit = limit
		}
	}

	var fn func(k *keeper.Keeper, ctx context.Context, req *types.QueryTokensRequest) (*types.QueryTokensResponse, error)
	switch {
	case req.CollectionAddr != "" && req.Owner == "" && req.TokenId == "":
		// query by collection only
		fn = getTokensByCollection
	case req.CollectionAddr != "" && req.Owner != "" && req.TokenId == "":
		// query by collection and owner
		fn = getTokensByCollectionAndOwner
	case req.CollectionAddr != "" && req.Owner == "" && req.TokenId != "":
		// query by collection and token id
		fn = getTokensByCollectionAndTokenId
	case req.CollectionAddr == "" && req.Owner != "" && req.TokenId == "":
		// query by owner only
		fn = getTokensByOwner
	case req.CollectionAddr != "" && req.Owner != "" && req.TokenId != "":
		// query by owner, collection and token id
		fn = getTokensByOwnerCollectionAndTokenId
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid query parameter")
	}

	return fn(q.Keeper, ctx, req)
}

// NewQuerier return new Querier instance
func NewQuerier(k *keeper.Keeper) Querier {
	return Querier{k}
}

func getCollectionNameFromPairSubmodule(ctx context.Context, collName string) (string, error) {
	return pair.GetPair(ctx, false, collName)
}

func getTokensByCollection(k *keeper.Keeper, ctx context.Context, req *types.QueryTokensRequest) (*types.QueryTokensResponse, error) {

	collAddr, err := getVMAddress(k.GetAddressCodec(), req.CollectionAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	colSdkAddr := getCosmosAddress(collAddr)

	res, pageRes, err := query.CollectionPaginate(ctx, tokenMap, req.Pagination,
		func(k collections.Pair[sdk.AccAddress, string], v types.IndexedToken) (*types.IndexedToken, error) {
			if slices.Equal(k.K1(), colSdkAddr) {
				return &v, nil
			}
			return nil, nil
		},
	)
	if err != nil {
		return nil, handleCollectionErr(err)
	}

	return &types.QueryTokensResponse{
		Tokens:     res,
		Pagination: pageRes,
	}, nil

}

func getTokensByCollectionAndOwner(k *keeper.Keeper, ctx context.Context, req *types.QueryTokensRequest) (*types.QueryTokensResponse, error) {
	collAddr, err := getVMAddress(k.GetAddressCodec(), req.CollectionAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	colSdkAddr := getCosmosAddress(collAddr)

	ownerAddr, err := getVMAddress(k.GetAddressCodec(), req.Owner)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	ownerSdkAddr := getCosmosAddress(ownerAddr)
	ownerSdkAddrStr := ownerSdkAddr.String()

	res, pageRes, err := query.CollectionPaginate(ctx, tokenMap, req.Pagination,
		func(k collections.Pair[sdk.AccAddress, string], v types.IndexedToken) (*types.IndexedToken, error) {
			if slices.Equal(k.K1(), colSdkAddr) && (v.OwnerAddr == ownerSdkAddrStr) {
				return &v, nil
			}
			return nil, nil
		},
	)
	if err != nil {
		return nil, handleCollectionErr(err)
	}
	if len(res) == 0 {
		return nil, status.Error(codes.NotFound, "tokens not found")
	}

	return &types.QueryTokensResponse{
		Tokens:     res,
		Pagination: pageRes,
	}, nil
}

func getTokensByCollectionAndTokenId(k *keeper.Keeper, ctx context.Context, req *types.QueryTokensRequest) (*types.QueryTokensResponse, error) {
	collAddr, err := getVMAddress(k.GetAddressCodec(), req.CollectionAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	colSdkAddr := getCosmosAddress(collAddr)

	token, err := tokenMap.Get(ctx, collections.Join(colSdkAddr, req.TokenId))
	if err != nil {
		return nil, handleCollectionErr(err)
	}

	return &types.QueryTokensResponse{
		Tokens: []*types.IndexedToken{&token},
	}, nil
}

func getTokensByOwner(k *keeper.Keeper, ctx context.Context, req *types.QueryTokensRequest) (*types.QueryTokensResponse, error) {

	ownerAddr, err := getVMAddress(k.GetAddressCodec(), req.Owner)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	ownerSdkAddr := getCosmosAddress(ownerAddr)
	ownerSdkAddrStr := ownerSdkAddr.String()

	store := k.GetStore()
	ownerStore := prefix.NewStore(*store, prefixTokenOwnerIndex)

	res, pageRes, err := query.GenericFilteredPaginate(
		k.GetCodec(),   /*codec*/
		ownerStore,     /* store */
		req.Pagination, /* page request */
		func(key []byte, val *types.IndexedToken) (*types.IndexedToken, error) {
			if val.OwnerAddr != ownerSdkAddrStr {
				return nil, nil
			}
			return val, nil
		}, /* onResult */
		func() *types.IndexedToken {
			return &types.IndexedToken{}
		}, /* constructor */
	)
	if err != nil {
		return nil, handleCollectionErr(err)
	}

	return &types.QueryTokensResponse{
		Tokens:     res,
		Pagination: pageRes,
	}, nil
}

func getTokensByOwnerCollectionAndTokenId(k *keeper.Keeper, ctx context.Context, req *types.QueryTokensRequest) (*types.QueryTokensResponse, error) {
	collAddr, err := getVMAddress(k.GetAddressCodec(), req.CollectionAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	colSdkAddr := getCosmosAddress(collAddr)

	token, err := tokenMap.Get(ctx, collections.Join(colSdkAddr, req.TokenId))
	if err != nil {
		return nil, handleCollectionErr(err)
	}

	if token.OwnerAddr != req.Owner {
		return nil, status.Error(codes.Unauthenticated, "invalid owner address")
	}

	return &types.QueryTokensResponse{
		Tokens: []*types.IndexedToken{&token},
	}, nil
}
