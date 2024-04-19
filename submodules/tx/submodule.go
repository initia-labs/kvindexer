package tx

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
	"github.com/initia-labs/kvindexer/submodules/tx/types"
	kvindexer "github.com/initia-labs/kvindexer/x/kvindexer/types"
)

var _ kvindexer.Submodule = TxSubmodule{}

type TxSubmodule struct {
	cdc codec.Codec

	sequence              *collections.Sequence
	txMap                 *collections.Map[string, sdk.TxResponse]
	txhashesBySequenceMap *collections.Map[uint64, string]
	txhashesByHeightMap   *collections.Map[collections.Pair[int64, uint64], string]
	txhashesByAccountMap  *collections.Map[collections.Pair[sdk.AccAddress, uint64], string]
	accountSequenceMap    *collections.Map[sdk.AccAddress, uint64]
}

func NewTxSubmodule(
	cdc codec.Codec,
	indexerKeeper collection.IndexerKeeper,
) (*TxSubmodule, error) {
	sequencePrefix := collection.NewPrefix(types.SubmoduleName, types.SequencePrefix)
	sequence, err := collection.AddSequence(indexerKeeper, sequencePrefix, "sequence")
	if err != nil {
		return nil, err
	}

	prefixTxs := collection.NewPrefix(types.SubmoduleName, types.TxsPrefix)
	txMap, err := collection.AddMap(indexerKeeper, prefixTxs, "txs", collections.StringKey, codec.CollValue[sdk.TxResponse](cdc))
	if err != nil {
		return nil, err
	}

	prefixTxsByAccount := collection.NewPrefix(types.SubmoduleName, types.TxsByAccountPrefix)
	txhashesByAccountMap, err := collection.AddMap(indexerKeeper, prefixTxsByAccount, "txs_by_account", collections.PairKeyCodec(sdk.AccAddressKey, collections.Uint64Key), collections.StringValue)
	if err != nil {
		return nil, err
	}

	prefixTxSequences := collection.NewPrefix(types.SubmoduleName, types.TxSequencePrefix)
	txhashesBySequenceMap, err := collection.AddMap(indexerKeeper, prefixTxSequences, "tx_sequences", collections.Uint64Key, collections.StringValue)
	if err != nil {
		return nil, err
	}

	prefixTxsByHeight := collection.NewPrefix(types.SubmoduleName, types.TxByHeightPrefix)
	txhashesByHeightMap, err := collection.AddMap(indexerKeeper, prefixTxsByHeight, "txs_by_height", collections.PairKeyCodec(collections.Int64Key, collections.Uint64Key), collections.StringValue)
	if err != nil {
		return nil, err
	}

	prefixAccountSequences := collection.NewPrefix(types.SubmoduleName, types.AccountSequencePrefix)
	accountSequenceMap, err := collection.AddMap(indexerKeeper, prefixAccountSequences, "account_sequences", sdk.AccAddressKey, collections.Uint64Value)
	if err != nil {
		return nil, err
	}

	return &TxSubmodule{
		cdc: cdc,

		sequence:              sequence,
		txMap:                 txMap,
		txhashesByAccountMap:  txhashesByAccountMap,
		txhashesBySequenceMap: txhashesBySequenceMap,
		txhashesByHeightMap:   txhashesByHeightMap,
		accountSequenceMap:    accountSequenceMap,
	}, nil
}

// Logger returns a module-specific logger.
func (sub TxSubmodule) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types.SubmoduleName)
}

func (sub TxSubmodule) Name() string {
	return types.SubmoduleName
}

func (sub TxSubmodule) Version() string {
	return types.Version
}

func (sub TxSubmodule) RegisterQueryHandlerClient(cc client.Context, mux *runtime.ServeMux) error {
	return types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(cc))
}

func (sub TxSubmodule) RegisterQueryServer(s grpc.Server) {
	types.RegisterQueryServer(s, NewQuerier(sub))
}

func (sub TxSubmodule) Prepare(ctx context.Context) error {
	return nil
}

func (sub TxSubmodule) Initialize(ctx context.Context) error {
	return nil
}

func (sub TxSubmodule) FinalizeBlock(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	return sub.finalizeBlock(ctx, req, res)
}

func (sub TxSubmodule) Commit(ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) error {
	return nil
}
