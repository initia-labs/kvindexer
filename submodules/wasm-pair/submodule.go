package wasm_pair

import (
	"context"

	"cosmossdk.io/collections"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/grpc"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/initia-labs/kvindexer/collection"
	pairtypes "github.com/initia-labs/kvindexer/pair/types"
	"github.com/initia-labs/kvindexer/submodules/wasm_pair/types"
	kvindexer "github.com/initia-labs/kvindexer/x/kvindexer/types"
)

var _ kvindexer.Submodule = PairSubmodule{}

type PairSubmodule struct {
	cdc codec.Codec

	channelKeeper  types.ChannelKeeper
	transferKeeper types.TransferKeeper

	nonFungiblePairsMap *collections.Map[string, string]
	fungiblePairsMap    *collections.Map[string, string]
}

func NewPairSubmodule(
	cdc codec.Codec,
	indexerKeeper collection.IndexerKeeper,
	channelKeeper types.ChannelKeeper,
	transferKeeper types.TransferKeeper,
) (*PairSubmodule, error) {
	prefixNonFungiblePairs := collection.NewPrefix(types.SubmoduleName, types.NonFungiblePairsPrefix)
	nonFungiblePairsMap, err := collection.AddMap(indexerKeeper, prefixNonFungiblePairs, "non_fungible_pairs", collections.StringKey, collections.StringValue)
	if err != nil {
		return nil, err
	}

	prefixFungiblePairs := collection.NewPrefix(types.SubmoduleName, types.FungiblePairsPrefix)
	fungiblePairsMap, err := collection.AddMap(indexerKeeper, prefixFungiblePairs, "fungible_pairs", collections.StringKey, collections.StringValue)
	if err != nil {
		return nil, err
	}

	return &PairSubmodule{
		cdc: cdc,

		channelKeeper:  channelKeeper,
		transferKeeper: transferKeeper,

		nonFungiblePairsMap: nonFungiblePairsMap,
		fungiblePairsMap:    fungiblePairsMap,
	}, nil
}

// Logger returns a module-specific logger.
func (sub PairSubmodule) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types.SubmoduleName)
}

func (sub PairSubmodule) Name() string {
	return types.SubmoduleName
}

func (sub PairSubmodule) Version() string {
	return types.Version
}

func (sub PairSubmodule) RegisterQueryHandlerClient(cc client.Context, mux *runtime.ServeMux) error {
	return pairtypes.RegisterQueryHandlerClient(context.Background(), mux, pairtypes.NewQueryClient(cc))
}

func (sub PairSubmodule) RegisterQueryServer(s grpc.Server) {
	pairtypes.RegisterQueryServer(s, NewQuerier(sub))
}

func (sub PairSubmodule) Prepare(ctx context.Context) error {
	return nil
}

func (sub PairSubmodule) Initialize(ctx context.Context) error {
	return nil
}

func (sub PairSubmodule) FinalizeBlock(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	return sub.finalizeBlock(ctx, req, res)
}

func (sub PairSubmodule) Commit(ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) error {
	return nil
}
