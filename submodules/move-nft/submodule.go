package move_nft

import (
	"context"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/grpc"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/initia-labs/kvindexer/collection"
	nfttypes "github.com/initia-labs/kvindexer/nft/types"
	"github.com/initia-labs/kvindexer/submodules/move-nft/types"
	kvindexer "github.com/initia-labs/kvindexer/x/kvindexer/types"
)

var _ kvindexer.Submodule = MoveNftSubmodule{}

type MoveNftSubmodule struct {
	ac  address.Codec
	cdc codec.Codec

	vmKeeper      types.MoveKeeper
	pairSubmodule types.PairSubmodule

	// collectionMap: key(collection address), value(collection)
	collectionMap *collections.Map[sdk.AccAddress, nfttypes.IndexedCollection]
	// collectionOwnerMap: key(owner address, collection address), value(owner's collection count)
	collectionOwnerMap *collections.Map[collections.Pair[sdk.AccAddress, sdk.AccAddress], uint64]
	// tokenMap: key(collection address, token id), value(token)
	tokenMap *collections.IndexedMap[collections.Pair[sdk.AccAddress, string], nfttypes.IndexedToken, TokenIndex]
	// tokenOwnerMap: key(owner address, collection address, token id), value(bool as placeholder)
	tokenOwnerMap *collections.Map[collections.Triple[sdk.AccAddress, sdk.AccAddress, string], bool]
}

func NewMoveNftSubmodule(
	ac address.Codec,
	cdc codec.Codec,
	indexerKeeper collection.IndexerKeeper,
	vmKeeper types.MoveKeeper,
	pairSubmodule types.PairSubmodule,
) (*MoveNftSubmodule, error) {
	collectionsPrefix := collection.NewPrefix(types.SubmoduleName, types.CollectionsPrefix)
	collectionMap, err := collection.AddMap(indexerKeeper, collectionsPrefix, "collections", sdk.AccAddressKey, codec.CollValue[nfttypes.IndexedCollection](cdc))
	if err != nil {
		return nil, err
	}

	collectionOwnersPrefix := collection.NewPrefix(types.SubmoduleName, types.CollectionOwnersPrefix)
	collectionOwnerMap, err := collection.AddMap(indexerKeeper, collectionOwnersPrefix, "collection_owners", collections.PairKeyCodec(sdk.AccAddressKey, sdk.AccAddressKey), collections.Uint64Value)
	if err != nil {
		return nil, err
	}

	tokensPrefix := collection.NewPrefix(types.SubmoduleName, types.TokensPrefix)
	tokenMap, err := collection.AddIndexedMap(indexerKeeper, tokensPrefix, "tokens", collections.PairKeyCodec(sdk.AccAddressKey, collections.StringKey), codec.CollValue[nfttypes.IndexedToken](cdc), newTokensIndex(ac, indexerKeeper))
	if err != nil {
		return nil, err
	}

	tokenOwnersPrefix := collection.NewPrefix(types.SubmoduleName, types.TokenOwnersPrefix)
	tokenOwnerMap, err := collection.AddMap(indexerKeeper, tokenOwnersPrefix, "token_owners", collections.TripleKeyCodec(sdk.AccAddressKey, sdk.AccAddressKey, collections.StringKey), collections.BoolValue)
	if err != nil {
		return nil, err
	}

	return &MoveNftSubmodule{
		ac:  ac,
		cdc: cdc,

		vmKeeper:      vmKeeper,
		pairSubmodule: pairSubmodule,

		collectionMap:      collectionMap,
		collectionOwnerMap: collectionOwnerMap,
		tokenMap:           tokenMap,
		tokenOwnerMap:      tokenOwnerMap,
	}, nil
}

// Logger returns a module-specific logger.
func (sub MoveNftSubmodule) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types.SubmoduleName)
}

func (sub MoveNftSubmodule) Name() string {
	return types.SubmoduleName
}

func (sub MoveNftSubmodule) Version() string {
	return types.Version
}

func (sub MoveNftSubmodule) RegisterQueryHandlerClient(cc client.Context, mux *runtime.ServeMux) error {
	return nfttypes.RegisterQueryHandlerClient(context.Background(), mux, nfttypes.NewQueryClient(cc))
}

func (sub MoveNftSubmodule) RegisterQueryServer(s grpc.Server) {
	nfttypes.RegisterQueryServer(s, NewQuerier(sub))
}

func (sub MoveNftSubmodule) Prepare(ctx context.Context) error {
	return nil
}

func (sub MoveNftSubmodule) Initialize(ctx context.Context) error {
	return nil
}

func (sub MoveNftSubmodule) FinalizeBlock(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	return sub.finalizeBlock(ctx, req, res)
}

func (sub MoveNftSubmodule) Commit(ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) error {
	return nil
}
