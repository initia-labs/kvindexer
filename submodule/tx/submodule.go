package tx

import (
	"context"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/initia-labs/kvindexer/module/keeper"
	"github.com/initia-labs/kvindexer/submodule/tx/types"
)

const submoduleName = "tx"

const txsByAccountPrefix = 0x10
const accountSequencePrefix = 0x20
const sequencePrefix = 0xa0
const txSequencePrefix = 0xb0
const txByHeght = 0xc0
const txsPrefix = 0xf0

const txsPerByAccountName = "txs_by_accounts"
const accountSequenceName = "account_sequence"
const txsName = "txs"
const txSequenceName = "tx_sequence"
const txByHeightName = "tx_by_height"
const sequenceName = "sequence"

var Version = "v0.0.1"

var (
	prefixTxsByAccount    = keeper.NewPrefix(submoduleName, txsByAccountPrefix)
	prefixAccountSequence = keeper.NewPrefix(submoduleName, accountSequencePrefix)
	prefixTxs             = keeper.NewPrefix(submoduleName, txsPrefix)
	prefixSequence        = keeper.NewPrefix(submoduleName, sequencePrefix)
	prefixTxSequence      = keeper.NewPrefix(submoduleName, txSequencePrefix)
	prefixTxByHeight      = keeper.NewPrefix(submoduleName, txByHeght)
)

// key: txhash value: tx
var txMap *collections.Map[string, sdk.TxResponse]

// global-sequence-for-every-tx
var sequence *collections.Sequence

// key: global-sequence-for-every-tx, value: txhash
var txhashesBySequence *collections.Map[uint64, string]

// key: [height, sequence-in-block], value: txhash
var txhashesByHeight *collections.Map[collections.Pair[int64, uint64], string]

// key: pair[account-address, sequence], value: tx
var txhashesByAccountMap *collections.Map[collections.Pair[sdk.AccAddress, uint64], string]

// key: account-address, value: sequence-per-account
var accountSequenceMap *collections.Map[sdk.AccAddress, uint64]

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
