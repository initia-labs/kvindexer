package block

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

	"github.com/initia-labs/initia/app/params"
	"github.com/initia-labs/kvindexer/collection"
	"github.com/initia-labs/kvindexer/submodules/block/types"
	kvindexer "github.com/initia-labs/kvindexer/x/kvindexer/types"
)

var _ kvindexer.Submodule = BlockSubmodule{}

type BlockSubmodule struct {
	encodingConfig params.EncodingConfig

	opChildKeeper types.OPChildKeeper

	blockByHeight *collections.Map[int64, types.Block]
}

func NewBlockSubmodule(
	encodingConfig params.EncodingConfig,
	indexerKeeper collection.IndexerKeeper,
	opChildKeeper types.OPChildKeeper,
) (*BlockSubmodule, error) {
	prefixBlock := collection.NewPrefix(types.SubmoduleName, types.BlockPrefix)
	blockByHeight, err := collection.AddMap(indexerKeeper, prefixBlock, "block_by_height", collections.Int64Key, codec.CollValue[types.Block](encodingConfig.Codec))
	if err != nil {
		return nil, err
	}

	return &BlockSubmodule{
		encodingConfig: encodingConfig,
		opChildKeeper:  opChildKeeper,
		blockByHeight:  blockByHeight,
	}, nil
}

// Logger returns a module-specific logger.
func (sub BlockSubmodule) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types.SubmoduleName)
}

func (sub BlockSubmodule) Name() string {
	return types.SubmoduleName
}

func (sub BlockSubmodule) Version() string {
	return types.Version
}

func (sub BlockSubmodule) RegisterQueryHandlerClient(cc client.Context, mux *runtime.ServeMux) error {
	return types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(cc))
}

func (sub BlockSubmodule) RegisterQueryServer(s grpc.Server) {
	types.RegisterQueryServer(s, NewQuerier(sub))
}

func (sub BlockSubmodule) Prepare(ctx context.Context) error {
	return nil
}

func (sub BlockSubmodule) Initialize(ctx context.Context) error {
	return nil
}

func (sub BlockSubmodule) FinalizeBlock(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	return sub.finalizeBlock(ctx, req, res)
}

func (sub BlockSubmodule) Commit(ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) error {
	return nil
}

func (sub BlockSubmodule) Prune(ctx context.Context, minHeight int64) error {
	return sub.prune(ctx, minHeight)
}
