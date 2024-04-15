package pair

import (
	"context"
	"sync"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/client"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/initia-labs/kvindexer/module/keeper"
	"github.com/initia-labs/kvindexer/submodule/pair/types"
)

const submoduleName = "pair"

const nonFungiblePairsPrefix = 0x10
const fungiblePairsPrefix = 0x11

const nonFungiblePairMapName = "nfpair"
const fungiblePairMapName = "fpair"

var Version = "v0.0.1"

var (
	prefixNonFungiblePairs = keeper.NewPrefix(submoduleName, nonFungiblePairsPrefix)
	prefixFungiblePairs    = keeper.NewPrefix(submoduleName, fungiblePairsPrefix)
)

// temporary storage for non-fungible token pairs from L2
// key: l2 collection address, value: l1 collection name
var nonFungiblePairsFromL2 sync.Map = sync.Map{}

// key: l2_collection_address, value: l1_collection_name
var nonFungiblePairsMap *collections.Map[string, string]

// key: l2_denom, value: l1_denom
var fungiblePairsMap *collections.Map[string, string]

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
