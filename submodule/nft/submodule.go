package nft

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/initia-labs/kvindexer/module/keeper"
	"github.com/initia-labs/kvindexer/submodule/nft/types"
)

const submoduleName = "nft"

var Version = "v0.0.1"

func RegisterQueryHandlerClient(cc client.Context, mux *runtime.ServeMux) error {
	return types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(cc))
}

func RegisterQueryServer(s grpc1.Server, k *keeper.Keeper) {
	types.RegisterQueryServer(s, NewQuerier(k))
}

var Submodule = keeper.Submodule{
	Name:                       submoduleName,
	Version:                    Version,
	Prepare:                    preparer,
	Initialize:                 initializer,
	HandleFinalizeBlock:        finalizeBlock,
	HandleCommit:               commit,
	RegisterQueryHandlerClient: RegisterQueryHandlerClient,
	RegisterQueryServer:        RegisterQueryServer,
}
