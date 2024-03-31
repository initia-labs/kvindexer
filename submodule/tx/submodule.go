package tx

import (
	"context"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/initia-labs/indexer/v2/module/keeper"
	"github.com/initia-labs/indexer/v2/submodule/tx/types"
)

const submoduleName = "tx"

const txsByAccountPrefix = 0x10
const accountSequencePrefix = 0x20
const txsPrefix = 0xf0

const txsPerByAccountName = "txs_by_accounts"
const accountSequenceName = "account_sequence"
const txsName = "txs"

var (
	prefixTxsByAccount    = keeper.NewPrefix(submoduleName, txsByAccountPrefix)
	prefixAccountSequence = keeper.NewPrefix(submoduleName, accountSequencePrefix)
	prefixTxs             = keeper.NewPrefix(submoduleName, txsPrefix)
)

// key: txhash value: tx
var txMap *collections.Map[string, sdk.TxResponse]

// key: pair[account-address, sequence], value: tx
var txhashesByAccountMap *collections.Map[collections.Pair[sdk.AccAddress, uint64], string]

// key: account-address, value: sequence
var accountSequenceMap *collections.Map[sdk.AccAddress, uint64]

func RegisterQueryHandlerClient(cc client.Context, mux *runtime.ServeMux) error {
	return types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(cc))
}

func RegisterQueryServer(s grpc1.Server, k *keeper.Keeper) {
	types.RegisterQueryServer(s, NewQuerier(k))
}

var Submodule = keeper.Submodule{
	Name:                       submoduleName,
	Prepare:                    preparer,
	Initialize:                 initializer,
	HandleFinalizeBlock:        finalizeBlock,
	HandleCommit:               commit,
	RegisterQueryHandlerClient: RegisterQueryHandlerClient,
	RegisterQueryServer:        RegisterQueryServer,
}
