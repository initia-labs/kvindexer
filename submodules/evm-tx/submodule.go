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
	"github.com/pkg/errors"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/initia-labs/kvindexer/collection"
	"github.com/initia-labs/kvindexer/submodules/evm-tx/types"
	kvindexer "github.com/initia-labs/kvindexer/x/kvindexer/types"
)

var _ kvindexer.Submodule = EvmTxSubmodule{}

type EvmTxSubmodule struct {
	cdc codec.Codec

	sequence              *collections.Sequence
	txMap                 *collections.Map[string, sdk.TxResponse]
	txhashesBySequenceMap *collections.Map[uint64, string]
	txhashesByHeightMap   *collections.Map[collections.Pair[int64, uint64], string]
	txhashesByAccountMap  *collections.Map[collections.Pair[sdk.AccAddress, uint64], string]
	accountSequenceMap    *collections.Map[sdk.AccAddress, uint64]

	oldSequence              *collections.Sequence
	oldTxMap                 *collections.Map[string, sdk.TxResponse]
	oldTxhashesBySequenceMap *collections.Map[uint64, string]
	oldTxhashesByHeightMap   *collections.Map[collections.Pair[int64, uint64], string]
	oldTxhashesByAccountMap  *collections.Map[collections.Pair[sdk.AccAddress, uint64], string]
	oldAccountSequenceMap    *collections.Map[sdk.AccAddress, uint64]

	// for pruning
	sequenceByHeightMap        *collections.Map[int64, uint64]
	accountSequenceByHeightMap *collections.Map[collections.Triple[int64, sdk.AccAddress, uint64], bool]

	// keeper
	//keeper collection.IndexerKeeper
}

func NewTxSubmodule(
	cdc codec.Codec,
	indexerKeeper collection.IndexerKeeper,
) (*EvmTxSubmodule, error) {
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

	prefixSequenceByHeight := collection.NewPrefix(types.SubmoduleName, types.SequenceByHeightPrefix)
	sequenceByHeightMap, err := collection.AddMap(indexerKeeper, prefixSequenceByHeight, "sequence_by_height", collections.Int64Key, collections.Uint64Value)
	if err != nil {
		return nil, err
	}

	prefixAccountSequenceByHeight := collection.NewPrefix(types.SubmoduleName, types.AccountSequenceByHeightPrefix)
	accountSequenceByHeightMap, err := collection.AddMap(indexerKeeper, prefixAccountSequenceByHeight, "account_sequence_by_height", collections.TripleKeyCodec(collections.Int64Key, sdk.AccAddressKey, collections.Uint64Key), collections.BoolValue)
	if err != nil {
		return nil, err
	}

	oldPrefix := collection.NewPrefix(oldModuleName, types.SequencePrefix)
	oldSeq, err := collection.AddSequence(indexerKeeper, oldPrefix, "osequence")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get old sequence")
	}

	oldPrefix = collection.NewPrefix(oldModuleName, types.TxsPrefix)
	oldTxMap, err := collection.AddMap(indexerKeeper, oldPrefix, "otxs", collections.StringKey, codec.CollValue[sdk.TxResponse](cdc))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get old tx map")
	}
	oldPrefix = collection.NewPrefix(oldModuleName, types.AccountSequencePrefix)
	oldAccountSequenceMap, err := collection.AddMap(indexerKeeper, oldPrefix, "oaccount_sequences", sdk.AccAddressKey, collections.Uint64Value)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get old accountSequence map")
	}
	oldPrefix = collection.NewPrefix(oldModuleName, types.TxsByAccountPrefix)
	oldTxhashesByAccountMap, err := collection.AddMap(indexerKeeper, oldPrefix, "otxs_by_account", collections.PairKeyCodec(sdk.AccAddressKey, collections.Uint64Key), collections.StringValue)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get old txhashedByAccount map")
	}

	oldPrefix = collection.NewPrefix(oldModuleName, types.TxSequencePrefix)
	oldTxhashesBySequenceMap, err := collection.AddMap(indexerKeeper, oldPrefix, "otx_sequences", collections.Uint64Key, collections.StringValue)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get old txhashesBySequence map")
	}

	oldPrefix = collection.NewPrefix(oldModuleName, types.TxByHeightPrefix)
	oldTxhashesByHeightMap, err := collection.AddMap(indexerKeeper, oldPrefix, "otxs_by_height", collections.PairKeyCodec(collections.Int64Key, collections.Uint64Key), collections.StringValue)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get old txhashesByHeight map")
	}

	/** no need
	oldPrefix = collection.NewPrefix(oldModuleName, types.SequenceByHeightPrefix)
	oldSequenceByHeightMap, err := collection.AddMap(s.keeper, oldPrefix, "sequence_by_height", collections.Int64Key, collections.Uint64Value)
	if err != nil {
		return errors.Wrap(err, "failed to get old sequenceByHeight map")
	}

	oldPrefix = collection.NewPrefix(oldModuleName, types.AccountSequenceByHeightPrefix)
	oldAccountSequenceByHeightMap, err := collection.AddMap(s.keeper, oldPrefix, "account_sequence_by_height", collections.TripleKeyCodec(collections.Int64Key, sdk.AccAddressKey, collections.Uint64Key), collections.BoolValue)
	if err != nil {
		return errors.Wrap(err, "failed to get old accountSequenceByHeight map")
	}
	*/

	return &EvmTxSubmodule{
		cdc: cdc,

		sequence:                   sequence,
		txMap:                      txMap,
		txhashesByAccountMap:       txhashesByAccountMap,
		txhashesBySequenceMap:      txhashesBySequenceMap,
		txhashesByHeightMap:        txhashesByHeightMap,
		accountSequenceMap:         accountSequenceMap,
		sequenceByHeightMap:        sequenceByHeightMap,
		accountSequenceByHeightMap: accountSequenceByHeightMap,

		// for patcher
		oldSequence:              oldSeq,
		oldTxMap:                 oldTxMap,
		oldTxhashesByAccountMap:  oldTxhashesByAccountMap,
		oldTxhashesBySequenceMap: oldTxhashesBySequenceMap,
		oldTxhashesByHeightMap:   oldTxhashesByHeightMap,
		oldAccountSequenceMap:    oldAccountSequenceMap,
		//keeper:                   indexerKeeper,
	}, nil
}

// Logger returns a module-specific logger.
func (sub EvmTxSubmodule) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types.SubmoduleName)
}

func (sub EvmTxSubmodule) Name() string {
	return types.SubmoduleName
}

func (sub EvmTxSubmodule) Version() string {
	return types.Version
}

func (sub EvmTxSubmodule) RegisterQueryHandlerClient(cc client.Context, mux *runtime.ServeMux) error {
	return types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(cc))
}

func (sub EvmTxSubmodule) RegisterQueryServer(s grpc.Server) {
	types.RegisterQueryServer(s, NewQuerier(sub))
}

func (sub EvmTxSubmodule) Prepare(ctx context.Context) error {
	return nil
}

func (sub EvmTxSubmodule) Initialize(ctx context.Context) error {
	return nil
}

func (sub EvmTxSubmodule) FinalizeBlock(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	if err := sub.PatchPrefix(ctx); err != nil {
		return err
	}
	return sub.finalizeBlock(ctx, req, res)
}

func (sub EvmTxSubmodule) Commit(ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) error {
	return nil
}

func (sub EvmTxSubmodule) Prune(ctx context.Context, minHeight int64) error {
	return sub.prune(ctx, minHeight)
}
