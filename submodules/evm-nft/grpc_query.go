package evm_nft

import (
	"context"
	"slices"

	"cosmossdk.io/collections"
	cosmoserr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	nfttypes "github.com/initia-labs/kvindexer/nft/types"
)

var _ nfttypes.QueryServer = (*Querier)(nil)

type Querier struct {
	EvmNFTSubmodule
}

func NewQuerier(mn EvmNFTSubmodule) nfttypes.QueryServer {
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

// Collection implements nfttypes.QueryServer.
func (q Querier) Collection(ctx context.Context, req *nfttypes.QueryCollectionRequest) (*nfttypes.QueryCollectionResponse, error) {

	collectionSdkAddr, err := getCosmosAddressFromString(q.ac, req.CollectionAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	collection, err := q.collectionMap.Get(ctx, collectionSdkAddr)
	if err != nil {
		return nil, handleCollectionErr(err)
	}
	collection.Collection.Name, _ = q.getCollectionNameFromPairSubmodule(ctx, collection.Collection.Name)

	return &nfttypes.QueryCollectionResponse{
		Collection: &collection,
	}, nil
}

// Collections implements nfttypes.QueryServer.
func (q Querier) CollectionsByAccount(ctx context.Context, req *nfttypes.QueryCollectionsByAccountRequest) (*nfttypes.QueryCollectionsResponse, error) {
	accountSdkAddr, err := getCosmosAddressFromString(q.ac, req.Account)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
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

	collections := []*nfttypes.IndexedCollection{}
	for _, collectionSdkAddr := range collectionSdkAddrs {
		collection, err := q.collectionMap.Get(ctx, collectionSdkAddr)
		if err != nil {
			return nil, handleCollectionErr(err)
		}
		collection.Collection.Name, _ = q.getCollectionNameFromPairSubmodule(ctx, collection.Collection.Name)
		collections = append(collections, &collection)
	}

	return &nfttypes.QueryCollectionsResponse{
		Collections: collections,
		Pagination:  pageRes,
	}, nil
}

// TokensByCollection implements nfttypes.QueryServer.
func (q Querier) TokensByCollection(ctx context.Context, req *nfttypes.QueryTokensByCollectionRequest) (*nfttypes.QueryTokensResponse, error) {
	if req.TokenId == "" {
		return q.getTokensByCollection(ctx, req)
	}
	return q.getTokensByCollectionAndTokenId(ctx, req)
}

// TokensByAccount implements nfttypes.QueryServer.
func (q Querier) TokensByAccount(ctx context.Context, req *nfttypes.QueryTokensByAccountRequest) (*nfttypes.QueryTokensResponse, error) {
	if req.CollectionAddr == "" {
		return q.getTokensByAccount(ctx, req)
	}
	if req.TokenId == "" {
		return q.getTokensByAccountAndCollection(ctx, req)
	}
	return q.getTokensByAccountCollectionAndTokenId(ctx, req)
}

func (sm EvmNFTSubmodule) getCollectionNameFromPairSubmodule(ctx context.Context, collName string) (string, error) {
	name, err := sm.pairSubmodule.GetPair(ctx, false, collName)
	if err != nil {
		return collName, err
	}

	return name, nil
}

func (sm EvmNFTSubmodule) getTokensByCollection(ctx context.Context, req *nfttypes.QueryTokensByCollectionRequest) (*nfttypes.QueryTokensResponse, error) {
	colSdkAddr, err := getCosmosAddressFromString(sm.ac, req.CollectionAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	res, pageRes, err := query.CollectionFilteredPaginate(ctx, sm.tokenMap, req.Pagination,
		func(key collections.Pair[sdk.AccAddress, string], v nfttypes.IndexedToken) (bool, error) {
			if slices.Equal(key.K1(), colSdkAddr) {
				return true, nil
			}
			return false, nil
		},
		func(k collections.Pair[sdk.AccAddress, string], v nfttypes.IndexedToken) (*nfttypes.IndexedToken, error) {
			v.CollectionName, _ = sm.getCollectionNameFromPairSubmodule(ctx, v.CollectionName)
			return &v, nil
		},
	)
	if err != nil {
		return nil, handleCollectionErr(err)
	}
	res = slices.DeleteFunc(res, func(item *nfttypes.IndexedToken) bool {
		return item == nil
	})
	res = slices.Clip(res)

	return &nfttypes.QueryTokensResponse{
		Tokens:     res,
		Pagination: pageRes,
	}, nil

}

func (sm EvmNFTSubmodule) getTokensByCollectionAndTokenId(ctx context.Context, req *nfttypes.QueryTokensByCollectionRequest) (*nfttypes.QueryTokensResponse, error) {
	colSdkAddr, err := getCosmosAddressFromString(sm.ac, req.CollectionAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	token, err := sm.tokenMap.Get(ctx, collections.Join(colSdkAddr, req.TokenId))
	if err != nil {
		return nil, handleCollectionErr(err)
	}
	token.CollectionName, _ = sm.getCollectionNameFromPairSubmodule(ctx, token.CollectionName)

	return &nfttypes.QueryTokensResponse{
		Tokens: []*nfttypes.IndexedToken{&token},
	}, nil
}

func (sm EvmNFTSubmodule) getTokensByAccount(ctx context.Context, req *nfttypes.QueryTokensByAccountRequest) (*nfttypes.QueryTokensResponse, error) {
	ownerSdkAddr, err := getCosmosAddressFromString(sm.ac, req.Account)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
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

	res := []*nfttypes.IndexedToken{}
	for _, identifier := range identifiers {
		token, err := sm.tokenMap.Get(ctx, identifier)
		if err != nil {
			return nil, handleCollectionErr(err)
		}
		token.CollectionName, _ = sm.getCollectionNameFromPairSubmodule(ctx, token.CollectionName)
		res = append(res, &token)
	}

	return &nfttypes.QueryTokensResponse{
		Tokens:     res,
		Pagination: pageRes,
	}, nil
}

func (sm EvmNFTSubmodule) getTokensByAccountAndCollection(ctx context.Context, req *nfttypes.QueryTokensByAccountRequest) (*nfttypes.QueryTokensResponse, error) {
	colSdkAddr, err := getCosmosAddressFromString(sm.ac, req.CollectionAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ownerSdkAddr, err := getCosmosAddressFromString(sm.ac, req.Account)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	ownerAddrStr := ownerSdkAddr.String()

	res, pageRes, err := query.CollectionFilteredPaginate(ctx, sm.tokenMap, req.Pagination,
		func(k collections.Pair[sdk.AccAddress, string], v nfttypes.IndexedToken) (bool, error) {
			if slices.Equal(k.K1(), colSdkAddr) && (v.OwnerAddr == ownerAddrStr) {
				return true, nil
			}
			return false, nil
		},
		func(k collections.Pair[sdk.AccAddress, string], v nfttypes.IndexedToken) (*nfttypes.IndexedToken, error) {
			v.CollectionName, _ = sm.getCollectionNameFromPairSubmodule(ctx, v.CollectionName)
			return &v, nil
		},
	)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	res = slices.DeleteFunc(res, func(item *nfttypes.IndexedToken) bool {
		return item == nil
	})
	res = slices.Clip(res)

	return &nfttypes.QueryTokensResponse{
		Tokens:     res,
		Pagination: pageRes,
	}, nil
}

func (sm EvmNFTSubmodule) getTokensByAccountCollectionAndTokenId(ctx context.Context, req *nfttypes.QueryTokensByAccountRequest) (*nfttypes.QueryTokensResponse, error) {
	colSdkAddr, err := getCosmosAddressFromString(sm.ac, req.CollectionAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	token, err := sm.tokenMap.Get(ctx, collections.Join(colSdkAddr, req.TokenId))
	if err != nil {
		return nil, handleCollectionErr(err)
	}

	if token.OwnerAddr != req.Account {
		return &nfttypes.QueryTokensResponse{
			Tokens: []*nfttypes.IndexedToken{},
		}, nil
	}

	token.CollectionName, _ = sm.getCollectionNameFromPairSubmodule(ctx, token.CollectionName)
	return &nfttypes.QueryTokensResponse{
		Tokens: []*nfttypes.IndexedToken{&token},
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
