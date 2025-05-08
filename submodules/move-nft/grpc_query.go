package move_nft

import (
	"context"
	"slices"
	"strings"

	"cosmossdk.io/collections"
	cosmoserr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	kvcollection "github.com/initia-labs/kvindexer/collection"
	nfttypes "github.com/initia-labs/kvindexer/nft/types"
	"github.com/initia-labs/kvindexer/util"
)

var _ nfttypes.QueryServer = (*Querier)(nil)

type Querier struct {
	MoveNftSubmodule
}

func NewQuerier(mn MoveNftSubmodule) nfttypes.QueryServer {
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

	return &nfttypes.QueryCollectionResponse{
		Collection: &collection,
	}, nil
}

// Collections implements nfttypes.QueryServer.
func (q Querier) Collections(ctx context.Context, req *nfttypes.QueryCollectionsRequest) (*nfttypes.QueryCollectionsResponse, error) {
	util.ValidatePageRequest(req.Pagination)

	collections, pageRes, err := query.CollectionPaginate(ctx, q.collectionMap, req.Pagination,
		func(k sdk.AccAddress, v nfttypes.IndexedCollection) (*nfttypes.IndexedCollection, error) {
			return &v, nil
		},
	)
	if err != nil {
		return nil, handleCollectionErr(err)
	}

	return &nfttypes.QueryCollectionsResponse{
		Collections: collections,
		Pagination:  pageRes,
	}, nil
}

// CollectionsByName implements nfttypes.QueryServer.
func (q Querier) CollectionsByName(ctx context.Context, req *nfttypes.QueryCollectionsByNameRequest) (*nfttypes.QueryCollectionsResponse, error) {
	util.ValidatePageRequest(req.Pagination)
	name := strings.ToLower(req.Name) // use lowercased name to support case insensitive search

	addrgrps, pageRes, err := query.CollectionPaginate(ctx, q.collectionNameMap, req.Pagination,
		func(k string, v string) (string, error) {
			return v, nil
		},
		func(opt *query.CollectionsPaginateOptions[string]) {
			opt.Prefix = &name
		},
	)
	if err != nil {
		return nil, handleCollectionErr(err)
	}
	colAddrs := expandString(addrgrps)
	collections := []*nfttypes.IndexedCollection{}
	for _, colAddr := range colAddrs {

		sdkAddr, err := sdk.AccAddressFromBech32(colAddr)
		if err != nil {
			q.Logger(ctx).Warn("invalid collection address found", "collection", colAddr, "action", "CollectionsByName", "error", err)
			continue
		}

		collection, err := q.collectionMap.Get(ctx, sdkAddr)
		if err != nil {
			q.Logger(ctx).Warn("index mismatch found", "collection", colAddr, "action", "CollectionsByName", "error", err)
			continue
		}
		collection.Collection.Name, _ = q.getCollectionNameFromPairSubmodule(ctx, collection.Collection.Name)
		collections = append(collections, &collection)
	}

	return &nfttypes.QueryCollectionsResponse{
		Collections: collections,
		Pagination:  pageRes,
	}, nil
}

// Collections implements nfttypes.QueryServer.
func (q Querier) CollectionsByAccount(ctx context.Context, req *nfttypes.QueryCollectionsByAccountRequest) (*nfttypes.QueryCollectionsResponse, error) {
	util.ValidatePageRequest(req.Pagination)
	accountAddr, err := getVMAddress(q.ac, req.Account)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	accountSdkAddr := getCosmosAddress(accountAddr)

	collectionSdkAddrs := []sdk.AccAddress{}
	_, pageRes, err := query.CollectionPaginate(ctx, q.collectionOwnerMap, req.Pagination,
		func(k collections.Pair[sdk.AccAddress, sdk.AccAddress], v uint64) (uint64, error) {
			collectionSdkAddrs = append(collectionSdkAddrs, k.K2())
			return v, nil
		},
		query.WithCollectionPaginationPairPrefix[sdk.AccAddress, sdk.AccAddress](accountSdkAddr),
	)
	if err != nil {
		return nil, handleCollectionErr(err)
	}

	indexedCollections := []*nfttypes.IndexedCollection{}
	for _, collectionSdkAddr := range collectionSdkAddrs {
		collection, err := q.collectionMap.Get(ctx, collectionSdkAddr)
		if err != nil {
			q.Logger(ctx).Warn("index mismatch found", "collection", collectionSdkAddr, "action", "CollectionsByAccount", "error", err)
			if cosmoserr.IsOf(err, collections.ErrNotFound) {
				pageRes.Total--
			}
			continue
		}
		collection.Collection.Name, _ = q.getCollectionNameFromPairSubmodule(ctx, collection.Collection.Name)
		indexedCollections = append(indexedCollections, &collection)
	}

	return &nfttypes.QueryCollectionsResponse{
		Collections: indexedCollections,
		Pagination:  pageRes,
	}, nil
}

// TokensByCollection implements nfttypes.QueryServer.
func (q Querier) TokensByCollection(ctx context.Context, req *nfttypes.QueryTokensByCollectionRequest) (*nfttypes.QueryTokensResponse, error) {
	util.ValidatePageRequest(req.Pagination)
	if req.TokenId == "" {
		return q.getTokensByCollection(ctx, req)
	}
	return q.getTokensByCollectionAndTokenId(ctx, req)
}

// TokensByAccount implements nfttypes.QueryServer.
func (q Querier) TokensByAccount(ctx context.Context, req *nfttypes.QueryTokensByAccountRequest) (*nfttypes.QueryTokensResponse, error) {
	util.ValidatePageRequest(req.Pagination)
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

func (sm MoveNftSubmodule) getTokensByCollection(ctx context.Context, req *nfttypes.QueryTokensByCollectionRequest) (*nfttypes.QueryTokensResponse, error) {
	collAddr, err := getVMAddress(sm.ac, req.CollectionAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	colSdkAddr := getCosmosAddress(collAddr)

	res, pageRes, err := query.CollectionPaginate(ctx, sm.tokenMap, req.Pagination,
		func(k collections.Pair[sdk.AccAddress, string], v nfttypes.IndexedToken) (*nfttypes.IndexedToken, error) {
			v.CollectionName, _ = sm.getCollectionNameFromPairSubmodule(ctx, v.CollectionName)
			return &v, nil
		},
		query.WithCollectionPaginationPairPrefix[sdk.AccAddress, string](colSdkAddr),
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

func (sm MoveNftSubmodule) getTokensByCollectionAndTokenId(ctx context.Context, req *nfttypes.QueryTokensByCollectionRequest) (*nfttypes.QueryTokensResponse, error) {
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

	return &nfttypes.QueryTokensResponse{
		Tokens: []*nfttypes.IndexedToken{&token},
	}, nil
}

func (sm MoveNftSubmodule) getTokensByAccount(ctx context.Context, req *nfttypes.QueryTokensByAccountRequest) (*nfttypes.QueryTokensResponse, error) {
	ownerAddr, err := getVMAddress(sm.ac, req.Account)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	ownerSdkAddr := getCosmosAddress(ownerAddr)
	identifiers := []collections.Pair[sdk.AccAddress, string]{}

	_, pageRes, err := query.CollectionPaginate(ctx, sm.tokenOwnerMap, req.Pagination,
		func(k collections.Triple[sdk.AccAddress, sdk.AccAddress, string], v bool) (bool, error) {
			identifiers = append(identifiers, collections.Join(k.K2(), k.K3()))
			return v, nil
		},
		kvcollection.WithCollectionPaginationTriplePrefix[sdk.AccAddress, sdk.AccAddress, string](ownerSdkAddr),
	)
	if err != nil {
		return nil, handleCollectionErr(err)
	}

	res := []*nfttypes.IndexedToken{}
	for _, identifier := range identifiers {
		token, err := sm.tokenMap.Get(ctx, identifier)
		if err != nil {
			sm.Logger(ctx).Warn("index mismatch found", "account", ownerSdkAddr, "action", "CollectionsByAccount", "error", err)
			if cosmoserr.IsOf(err, collections.ErrNotFound) {
				pageRes.Total--
			}
			continue
		}
		token.CollectionName, _ = sm.getCollectionNameFromPairSubmodule(ctx, token.CollectionName)
		res = append(res, &token)
	}

	return &nfttypes.QueryTokensResponse{
		Tokens:     res,
		Pagination: pageRes,
	}, nil
}

func (sm MoveNftSubmodule) getTokensByAccountAndCollection(ctx context.Context, req *nfttypes.QueryTokensByAccountRequest) (*nfttypes.QueryTokensResponse, error) {
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

	identifiers := []collections.Pair[sdk.AccAddress, string]{}
	_, pageRes, err := query.CollectionPaginate(ctx, sm.tokenOwnerMap, req.Pagination,
		func(k collections.Triple[sdk.AccAddress, sdk.AccAddress, string], v bool) (bool, error) {
			identifiers = append(identifiers, collections.Join(k.K2(), k.K3()))
			return v, nil
		},
		kvcollection.WithCollectionPaginationTriplePrefix2[sdk.AccAddress, sdk.AccAddress, string](ownerSdkAddr, colSdkAddr),
	)
	if err != nil {
		return nil, handleCollectionErr(err)
	}
	res := []*nfttypes.IndexedToken{}
	for _, identifier := range identifiers {
		token, err := sm.tokenMap.Get(ctx, identifier)
		if err != nil {
			sm.Logger(ctx).Warn("index mismatch found", "account", ownerSdkAddr, "collection", colSdkAddr, "action", "GetTokensByAccountAndCollection", "error", err)
			if cosmoserr.IsOf(err, collections.ErrNotFound) {
				pageRes.Total--
			}
			continue
		}
		token.CollectionName, _ = sm.getCollectionNameFromPairSubmodule(ctx, token.CollectionName)
		res = append(res, &token)
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

func (sm MoveNftSubmodule) getTokensByAccountCollectionAndTokenId(ctx context.Context, req *nfttypes.QueryTokensByAccountRequest) (*nfttypes.QueryTokensResponse, error) {
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
		return &nfttypes.QueryTokensResponse{
			Tokens: []*nfttypes.IndexedToken{},
		}, nil
	}

	token.CollectionName, _ = sm.getCollectionNameFromPairSubmodule(ctx, token.CollectionName)
	return &nfttypes.QueryTokensResponse{
		Tokens: []*nfttypes.IndexedToken{&token},
	}, nil
}
