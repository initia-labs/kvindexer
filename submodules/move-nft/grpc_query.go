package move_nft

import (
	"context"
	"slices"

	"cosmossdk.io/collections"
	cosmoserr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/initia-labs/kvindexer/submodules/move-nft/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = (*Querier)(nil)

type Querier struct {
	MoveNftSubmodule
}

func NewQuerier(mn MoveNftSubmodule) types.QueryServer {
	return Querier{mn}
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
	collectionAddr, err := getVMAddress(q.ac, req.CollectionAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	collectionSdkAddr := getCosmosAddress(collectionAddr)

	collection, err := q.collectionMap.Get(ctx, collectionSdkAddr)
	if err != nil {
		return nil, handleCollectionErr(err)
	}
	collection.Collection.Name, _ = q.getCollectionNameFromPairSubmodule(ctx, collection.Collection.Name)

	return &types.QueryCollectionResponse{
		Collection: &collection,
	}, nil
}

// Collections implements types.QueryServer.
func (q Querier) CollectionsByAccount(ctx context.Context, req *types.QueryCollectionsByAccountRequest) (*types.QueryCollectionsResponse, error) {
	accountAddr, err := getVMAddress(q.ac, req.Account)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	accountSdkAddr := getCosmosAddress(accountAddr)
	accountAddrString := accountSdkAddr.String()

	collectionSdkAddrs := []sdk.AccAddress{}
	_, pageRes, err := query.CollectionFilteredPaginate(ctx, q.collectionOwnerMap, req.Pagination,
		func(k collections.Pair[sdk.AccAddress, sdk.AccAddress], v uint64) (bool, error) {
			if k.K1().String() == accountAddrString {
				return true, nil
			}
			return false, nil
		},
		func(k collections.Pair[sdk.AccAddress, sdk.AccAddress], v uint64) (uint64, error) {
			collectionSdkAddrs = append(collectionSdkAddrs, k.K2())
			return v, nil
		},
		query.WithCollectionPaginationPairPrefix[sdk.AccAddress, sdk.AccAddress](accountSdkAddr),
	)
	if err != nil {
		return nil, handleCollectionErr(err)
	}

	collections := []*types.IndexedCollection{}
	for _, collectionSdkAddr := range collectionSdkAddrs {
		collection, err := q.collectionMap.Get(ctx, collectionSdkAddr)
		if err != nil {
			return nil, handleCollectionErr(err)
		}
		collection.Collection.Name, _ = q.getCollectionNameFromPairSubmodule(ctx, collection.Collection.Name)
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
		return q.getTokensByCollection(ctx, req)
	}
	return q.getTokensByCollectionAndTokenId(ctx, req)
}

// TokensByAccount implements types.QueryServer.
func (q Querier) TokensByAccount(ctx context.Context, req *types.QueryTokensByAccountRequest) (*types.QueryTokensResponse, error) {
	if req.CollectionAddr == "" {
		return q.getTokensByAccount(ctx, req)
	}
	if req.TokenId == "" {
		return q.getTokensByAccountAndCollection(ctx, req)
	}
	return q.getTokensByAccountCollectionAndTokenId(ctx, req)
}

func (sm MoveNftSubmodule) getCollectionNameFromPairSubmodule(ctx context.Context, collName string) (string, error) {
	name, err := sm.pairSubmodule.GetPair(ctx, false, collName)
	if err != nil {
		return collName, err
	}

	return name, nil
}

func (sm MoveNftSubmodule) getTokensByCollection(ctx context.Context, req *types.QueryTokensByCollectionRequest) (*types.QueryTokensResponse, error) {
	collAddr, err := getVMAddress(sm.ac, req.CollectionAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	colSdkAddr := getCosmosAddress(collAddr)

	res, pageRes, err := query.CollectionFilteredPaginate(ctx, sm.tokenMap, req.Pagination,
		func(key collections.Pair[sdk.AccAddress, string], v types.IndexedToken) (bool, error) {
			if slices.Equal(key.K1(), colSdkAddr) {
				return true, nil
			}
			return false, nil
		},
		func(k collections.Pair[sdk.AccAddress, string], v types.IndexedToken) (*types.IndexedToken, error) {
			v.CollectionName, _ = sm.getCollectionNameFromPairSubmodule(ctx, v.CollectionName)
			return &v, nil
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

func (sm MoveNftSubmodule) getTokensByCollectionAndTokenId(ctx context.Context, req *types.QueryTokensByCollectionRequest) (*types.QueryTokensResponse, error) {
	collAddr, err := getVMAddress(sm.ac, req.CollectionAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	colSdkAddr := getCosmosAddress(collAddr)

	token, err := sm.tokenMap.Get(ctx, collections.Join(colSdkAddr, req.TokenId))
	if err != nil {
		return nil, handleCollectionErr(err)
	}
	token.CollectionName, _ = sm.getCollectionNameFromPairSubmodule(ctx, token.CollectionName)

	return &types.QueryTokensResponse{
		Tokens: []*types.IndexedToken{&token},
	}, nil
}

func (sm MoveNftSubmodule) getTokensByAccount(ctx context.Context, req *types.QueryTokensByAccountRequest) (*types.QueryTokensResponse, error) {
	ownerAddr, err := getVMAddress(sm.ac, req.Account)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	ownerSdkAddr := getCosmosAddress(ownerAddr)
	identifiers := []collections.Pair[sdk.AccAddress, string]{}

	_, pageRes, err := query.CollectionFilteredPaginate(ctx, sm.tokenOwnerMap, req.Pagination,
		func(k collections.Triple[sdk.AccAddress, sdk.AccAddress, string], _ bool) (bool, error) {
			return true, nil
		},
		func(k collections.Triple[sdk.AccAddress, sdk.AccAddress, string], v bool) (bool, error) {
			identifiers = append(identifiers, collections.Join(k.K2(), k.K3()))
			return v, nil
		},
		WithCollectionPaginationTriplePrefix[sdk.AccAddress, sdk.AccAddress, string](ownerSdkAddr),
	)
	if err != nil {
		return nil, handleCollectionErr(err)
	}

	res := []*types.IndexedToken{}
	for _, identifier := range identifiers {
		token, err := sm.tokenMap.Get(ctx, identifier)
		if err != nil {
			return nil, handleCollectionErr(err)
		}
		token.CollectionName, _ = sm.getCollectionNameFromPairSubmodule(ctx, token.CollectionName)
		res = append(res, &token)
	}

	return &types.QueryTokensResponse{
		Tokens:     res,
		Pagination: pageRes,
	}, nil
}

func (sm MoveNftSubmodule) getTokensByAccountAndCollection(ctx context.Context, req *types.QueryTokensByAccountRequest) (*types.QueryTokensResponse, error) {
	collAddr, err := getVMAddress(sm.ac, req.CollectionAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	colSdkAddr := getCosmosAddress(collAddr)

	ownerAddr, err := getVMAddress(sm.ac, req.Account)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	ownerSdkAddr := getCosmosAddress(ownerAddr)
	ownerAddrStr := ownerSdkAddr.String()

	res, pageRes, err := query.CollectionPaginate(ctx, sm.tokenMap, req.Pagination,
		func(k collections.Pair[sdk.AccAddress, string], v types.IndexedToken) (*types.IndexedToken, error) {
			if slices.Equal(k.K1(), colSdkAddr) && (v.OwnerAddr == ownerAddrStr) {
				v.CollectionName, _ = sm.getCollectionNameFromPairSubmodule(ctx, v.CollectionName)
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

func (sm MoveNftSubmodule) getTokensByAccountCollectionAndTokenId(ctx context.Context, req *types.QueryTokensByAccountRequest) (*types.QueryTokensResponse, error) {
	collAddr, err := getVMAddress(sm.ac, req.CollectionAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	colSdkAddr := getCosmosAddress(collAddr)

	token, err := sm.tokenMap.Get(ctx, collections.Join(colSdkAddr, req.TokenId))
	if err != nil {
		return nil, handleCollectionErr(err)
	}

	if token.OwnerAddr != req.Account {
		return &types.QueryTokensResponse{
			Tokens: []*types.IndexedToken{},
		}, nil
	}

	token.CollectionName, _ = sm.getCollectionNameFromPairSubmodule(ctx, token.CollectionName)
	return &types.QueryTokensResponse{
		Tokens: []*types.IndexedToken{&token},
	}, nil
}

// WithCollectionPaginationTriplePrefix applies a prefix to a collection, whose key is a collection.Triple,
// being paginated that needs prefixing.
func WithCollectionPaginationTriplePrefix[K1, K2, K3 any](prefix K1) func(o *query.CollectionsPaginateOptions[collections.Triple[K1, K2, K3]]) {
	return func(o *query.CollectionsPaginateOptions[collections.Triple[K1, K2, K3]]) {
		prefix := collections.TriplePrefix[K1, K2, K3](prefix)
		o.Prefix = &prefix
	}
}
