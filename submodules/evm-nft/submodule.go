package evm_nft

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

	"github.com/initia-labs/initia/app/params"
	evmkeeper "github.com/initia-labs/minievm/x/evm/keeper"

	"github.com/initia-labs/kvindexer/collection"
	nfttypes "github.com/initia-labs/kvindexer/nft/types"
	"github.com/initia-labs/kvindexer/submodules/evm-nft/types"
	kvindexer "github.com/initia-labs/kvindexer/x/kvindexer/types"
)

var _ kvindexer.Submodule = EvmNFTSubmodule{}

type EvmNFTSubmodule struct {
	ac             address.Codec
	encodingConfig params.EncodingConfig

	vmKeeper      *evmkeeper.Keeper
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

func NewEvmNFTSubmodule(
	ac address.Codec,
	encodingConfig params.EncodingConfig,
	indexerKeeper collection.IndexerKeeper,
	vmKeeper *evmkeeper.Keeper,
	pairSubmodule types.PairSubmodule,
) (*EvmNFTSubmodule, error) {
	cdc := encodingConfig.Codec

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

	return &EvmNFTSubmodule{
		ac:             ac,
		encodingConfig: encodingConfig,

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
func (sub EvmNFTSubmodule) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types.SubmoduleName)
}

func (sub EvmNFTSubmodule) Name() string {
	return types.SubmoduleName
}

func (sub EvmNFTSubmodule) Version() string {
	return types.Version
}

func (sub EvmNFTSubmodule) RegisterQueryHandlerClient(cc client.Context, mux *runtime.ServeMux) error {
	return nfttypes.RegisterQueryHandlerClient(context.Background(), mux, nfttypes.NewQueryClient(cc))
}

func (sub EvmNFTSubmodule) RegisterQueryServer(s grpc.Server) {
	nfttypes.RegisterQueryServer(s, NewQuerier(sub))
}

func (sub EvmNFTSubmodule) Prepare(ctx context.Context) error {
	return nil
}

func (sub EvmNFTSubmodule) Initialize(ctx context.Context) error {
	return nil
}

func (sub EvmNFTSubmodule) FinalizeBlock(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	if err := sub.migrateHandler(ctx); err != nil {
		sub.Logger(ctx).Error("failed to migrate", "error", err)
		return err
	}
	return sub.finalizeBlock(ctx, req, res)
}

func (sub EvmNFTSubmodule) Commit(ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) error {
	return nil
}

func (sub EvmNFTSubmodule) Prune(ctx context.Context, minHeight int64) error {
	return nil
}
