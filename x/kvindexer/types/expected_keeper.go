package types

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/gogoproto/grpc"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

// Submodule is an interface that defines the methods that a submodule must implement.
type Submodule interface {
	Prepare(ctx context.Context) error
	Initialize(ctx context.Context) error
	FinalizeBlock(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error
	Commit(ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) error
	RegisterQueryHandlerClient(ctx client.Context, mux *runtime.ServeMux) error
	RegisterQueryServer(s grpc.Server)
	Prune(ctx context.Context, minHeight int64) error

	Name() string
	Version() string
}
