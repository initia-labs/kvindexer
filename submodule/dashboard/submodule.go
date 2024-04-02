package dashboard

import (
	"context"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/client"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/initia-labs/kvindexer/module/keeper"
	"github.com/initia-labs/kvindexer/submodule/dashboard/types"
)

const submoduleName = "dashboard"

const newAccsPrefix = 0x10
const newAccsCountPrefix = 0x11
const txCountPrefix = 0x20
const supplyPrefix = 0x21
const totalAccsCountPrefix = 0x30
const lastAccountNumberPrefix = 0x40

const newAccsByHeightName = "new_accounts_by_height"
const newAccsCountByDateName = "new_accounts_count_by_date"
const txCountByDateKeyName = "tx_count_by_date"
const supplyByDateKeyName = "supply_by_date"
const totalAccsCountByDateName = "total_account_count_by_date"
const lastAccountNumberName = "last_account_number"

var (
	prefixAccountMapByHeight      = keeper.NewPrefix(submoduleName, newAccsPrefix)
	prefixAccountCountMapByDate   = keeper.NewPrefix(submoduleName, newAccsCountPrefix)
	prefixTxCountByDate           = keeper.NewPrefix(submoduleName, txCountPrefix)
	prefixSupplyByDate            = keeper.NewPrefix(submoduleName, supplyPrefix)
	prefixTotalAccountCountByDate = keeper.NewPrefix(submoduleName, totalAccsCountPrefix)
	prefixLastAccountNumber       = keeper.NewPrefix(submoduleName, lastAccountNumberPrefix)
)

// key: height, value: address list seperated by comma
var accountMapByHeight *collections.Map[int64, string]

// key: date string, value: new account count
var newAccountCountMapByDate *collections.Map[string, uint64]

// key: date string, value: total account count
var totalAccountBaseCountByDate *collections.Map[string, uint64]

// value: last account number: should be same with `k.AccountKeeper.AccountNumber.Peek(ctx)`
var lastAccountNumber *collections.Item[uint64]

// key: date string, value: accumulative tx count
var txCountByDate *collections.Map[string, uint64]

// key: date string, value: total supply []byte
var supplyByDate *collections.Map[string, uint64]

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
