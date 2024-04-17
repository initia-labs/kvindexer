package keeper

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	grpc1 "github.com/cosmos/gogoproto/grpc"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	storetypes "cosmossdk.io/store/types"
)

type Submodule struct {
	// Name must be unique
	Name string
	// Version is the version of the submodule
	Version string
	//	RequiredKeys are the keys that are required to be present in changeSet
	RequiredKeys map[string]*storetypes.StoreKey
	// Prepare is a function that will be called when the submodule is prepared
	Prepare Preparer
	// Initializer is a function that will be called when the submodule is started
	Initialize Initializer
	// FinalizeBlockHandler is a function that will be called when the block is finalized
	HandleFinalizeBlock FinalizeBlockHandler
	// CommitHandler is a function that will be called when the block is committed
	HandleCommit CommitHandler
	// RegisterQueryHandlerClient Func
	RegisterQueryHandlerClient RegisterQueryHandlerClientFunc
	// RegisterQueryServer is a function that will be called when the query server is registered
	RegisterQueryServer RegisterQueryServerFunc
}

// NOTE: 'ctx' does NOT contain sdk context!
type Preparer func(keeper *Keeper, ctx context.Context) error

// NOTE: 'ctx' does NOT contain sdk context!
type Initializer func(keeper *Keeper, ctx context.Context) error

type FinalizeBlockHandler func(keeper *Keeper, ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error

type CommitHandler func(keeper *Keeper, ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) error

type RegisterQueryHandlerClientFunc func(ctx client.Context, mux *runtime.ServeMux) error
type RegisterQueryServerFunc func(s grpc1.Server, k *Keeper)
