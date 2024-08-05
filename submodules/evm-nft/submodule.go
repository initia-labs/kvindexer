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

	evmkeeper "github.com/initia-labs/minievm/x/evm/keeper"

	"github.com/initia-labs/kvindexer/collection"
	nfttypes "github.com/initia-labs/kvindexer/nft/types"
	"github.com/initia-labs/kvindexer/submodules/evm-nft/types"
	kvindexer "github.com/initia-labs/kvindexer/x/kvindexer/types"
)

var _ kvindexer.Submodule = EvmNFTSubmodule{}

type EvmNFTSubmodule struct {
	ac  address.Codec
	cdc codec.Codec

	vmKeeper      *evmkeeper.Keeper
	pairSubmodule types.PairSubmodule

	// collectionMap: key(collection address), value(collection)
	collectionMap *collections.Map[sdk.AccAddress, nfttypes.IndexedCollection]
	// collectionOwnerMap: key(owner address, collection address), value(collection`s object address)
	collectionOwnerMap *collections.Map[collections.Pair[sdk.AccAddress, sdk.AccAddress], uint64]
	// tokenMap: key(collection address, token id), value(token)
	tokenMap *collections.Map[collections.Pair[sdk.AccAddress, string], nfttypes.IndexedToken]
	// tokenOwnerMap: key(owner address, collection address, token id), value(bool as placeholder)
	tokenOwnerMap *collections.Map[collections.Triple[sdk.AccAddress, sdk.AccAddress, string], bool]
}

func NewevmNFTSubmodule(
	ac address.Codec,
	cdc codec.Codec,
	indexerKeeper collection.IndexerKeeper,
	vmKeeper *evmkeeper.Keeper,
	pairSubmodule types.PairSubmodule,
) (*EvmNFTSubmodule, error) {
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
	tokenMap, err := collection.AddMap(indexerKeeper, tokensPrefix, "tokens", collections.PairKeyCodec(sdk.AccAddressKey, collections.StringKey), codec.CollValue[nfttypes.IndexedToken](cdc))
	if err != nil {
		return nil, err
	}

	tokenOwnersPrefix := collection.NewPrefix(types.SubmoduleName, types.TokenOwnersPrefix)
	tokenOwnerMap, err := collection.AddMap(indexerKeeper, tokenOwnersPrefix, "token_owners", collections.TripleKeyCodec(sdk.AccAddressKey, sdk.AccAddressKey, collections.StringKey), collections.BoolValue)
	if err != nil {
		return nil, err
	}

	return &EvmNFTSubmodule{
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
	return sub.finalizeBlock(ctx, req, res)
}

func (sub EvmNFTSubmodule) Commit(ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) error {
	return nil
}
