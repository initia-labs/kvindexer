package kvindexer

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/initia-labs/kvindexer/x/kvindexer/keeper"
	"github.com/initia-labs/kvindexer/x/kvindexer/types"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ module.HasServices    = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the move module.
type AppModuleBasic struct {
	//cdc     codec.Codec
	keeper *keeper.Keeper
}

func (b AppModuleBasic) RegisterLegacyAminoCodec(amino *codec.LegacyAmino) { //nolint:staticcheck
	/*nop*/ //types.RegisterLegacyAminoCodec(amino)
}

func (b AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, serveMux *runtime.ServeMux) {
	err := types.RegisterQueryHandlerClient(context.Background(), serveMux, types.NewQueryClient(clientCtx))
	if err != nil {
		panic(err)
	}

	submodules := b.keeper.GetSubmodules()
	for _, sm := range submodules {
		err := sm.RegisterQueryHandlerClient(clientCtx, serveMux)
		if err != nil {
			panic(err)
		}
	}
}

// RegisterInterfaces implements InterfaceModule
func (b AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	/*nop*/ //types.RegisterInterfaces(registry)
}

// Name returns the move module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// AppModule implements an application module for the move module.
// Normally AppModule has this method, not AppModuleBasic, but indexer module has this method in AppModuleBasic
// because indexer module is not a real module and don't related to the consensus.
func (am AppModuleBasic) RegisterServices(cfg module.Configurator) {
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQuerier(am.keeper))

	submodules := am.keeper.GetSubmodules()
	for _, sm := range submodules {
		sm.RegisterQueryServer(cfg.QueryServer())
	}
}

// NewAppModuleBasic creates AppModuleBasic
func NewAppModuleBasic(
	keeper *keeper.Keeper,
) AppModuleBasic {
	return AppModuleBasic{
		keeper: keeper,
	}
}
