package wasm_nft

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
	"github.com/initia-labs/kvindexer/submodules/wasm-nft/types"
	kvindexer "github.com/initia-labs/kvindexer/x/kvindexer/types"
)

var _ kvindexer.Submodule = WasmNFTSubmodule{}

type WasmNFTSubmodule struct {
	ac  address.Codec
	cdc codec.Codec

	vmKeeper      types.WasmKeeper
	pairSubmodule types.PairSubmodule

	// collectionMap: key(collection address), value(collection)
	collectionMap *collections.Map[sdk.AccAddress, nfttypes.IndexedCollection]
	// collectionOwnerMap: key(owner address, collection address), value(owner's collection count)
	collectionOwnerMap *collections.Map[collections.Pair[sdk.AccAddress, sdk.AccAddress], uint64]
	// collectionNameMap: key(collection name), value(collection address)
	collectionNameMap *collections.Map[string, string]
	// tokenMap: key(collection address, token id), value(token)
	tokenMap *collections.Map[collections.Pair[sdk.AccAddress, string], nfttypes.IndexedToken]
	// tokenOwnerMap: key(owner address, collection address, token id), value(bool as placeholder)
	tokenOwnerMap *collections.Map[collections.Triple[sdk.AccAddress, sdk.AccAddress, string], bool]
	// migrationInfo stores json and internal use only
	migrationInfo *collections.Map[string, string]
}

func NewWasmNFTSubmodule(
	ac address.Codec,
	cdc codec.Codec,
	indexerKeeper collection.IndexerKeeper,
	vmKeeper types.WasmKeeper,
	pairSubmodule types.PairSubmodule,
) (*WasmNFTSubmodule, error) {
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

	collectionNamesPrefix := collection.NewPrefix(types.SubmoduleName, types.CollectionNamesPrefix)
	collectionNameMap, err := collection.AddMap(indexerKeeper, collectionNamesPrefix, "collection_names", collections.StringKey, collections.StringValue)
	if err != nil {
		return nil, err
	}

	tokensPrefix := collection.NewPrefix(types.SubmoduleName, types.TokensPrefix)
	tokenMap, err := collection.AddMap(indexerKeeper, tokensPrefix, "tokens", collections.PairKeyCodec(sdk.AccAddressKey, collections.StringKey), codec.CollValue[nfttypes.IndexedToken](cdc))
	if err != nil {
		return nil, err
	}

	tokenOwnersPrefix := collection.NewPrefix(types.SubmoduleName, types.TokenOwnersPrefix)
	tokenOwnerMap, err := collection.AddMap(indexerKeeper, tokenOwnersPrefix, "token_owners", collections.TripleKeyCodec(sdk.AccAddressKey, sdk.AccAddressKey, collections.StringKey), collections.BoolValue)
	if err != nil {
		return nil, err
	}

	migrationPrefix := collection.NewPrefix(string(types.SubmoduleName), types.MigrationPrefix)
	migrationMap, err := collection.AddMap(indexerKeeper, migrationPrefix, "migration", collections.StringKey, collections.StringValue)
	if err != nil {
		return nil, err
	}

	return &WasmNFTSubmodule{
		ac:  ac,
		cdc: cdc,

		vmKeeper:      vmKeeper,
		pairSubmodule: pairSubmodule,

		collectionMap:      collectionMap,
		collectionOwnerMap: collectionOwnerMap,
		collectionNameMap:  collectionNameMap,
		tokenMap:           tokenMap,
		tokenOwnerMap:      tokenOwnerMap,
		migrationInfo:      migrationMap,
	}, nil
}

// Logger returns a module-specific logger.
func (sm WasmNFTSubmodule) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types.SubmoduleName)
}

func (sm WasmNFTSubmodule) Name() string {
	return types.SubmoduleName
}

func (sm WasmNFTSubmodule) Version() string {
	return types.Version
}

func (sm WasmNFTSubmodule) RegisterQueryHandlerClient(cc client.Context, mux *runtime.ServeMux) error {
	return nfttypes.RegisterQueryHandlerClient(context.Background(), mux, nfttypes.NewQueryClient(cc))
}

func (sm WasmNFTSubmodule) RegisterQueryServer(s grpc.Server) {
	nfttypes.RegisterQueryServer(s, NewQuerier(sm))
}

func (sm WasmNFTSubmodule) Prepare(ctx context.Context) error {
	return nil
}

func (sm WasmNFTSubmodule) Initialize(ctx context.Context) error {
	return nil
}

func (sm WasmNFTSubmodule) FinalizeBlock(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	if err := sm.migrateHandler(ctx); err != nil {
		sm.Logger(ctx).Error("failed to migrate", "error", err)
		return err
	}
	return sm.finalizeBlock(ctx, req, res)
}

func (sm WasmNFTSubmodule) Commit(ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) error {
	return nil
}

func (sub WasmNFTSubmodule) Prune(ctx context.Context, minHeight int64) error {
	return nil
}
