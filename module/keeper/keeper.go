package keeper

import (
	"context"
	"errors"
	"path"

	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestoretypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/store/dbadapter"
	storetypes "cosmossdk.io/store/types"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"

	"github.com/initia-labs/kvindexer/config"
	"github.com/initia-labs/kvindexer/module/types"
	"github.com/initia-labs/kvindexer/store"
)

const StoreName = "indexer"
const DataDir = "data"

type Keeper struct {
	cdc   codec.Codec
	store storetypes.CacheKVStore
	//store store.CosmosKVStore

	// used only for staking feature
	DistrKeeper         types.DistributionKeeper
	StakingKeeper       types.StakingKeeper
	RewardKeeper        types.RewardKeeper
	CommunityPoolKeeper types.CommunityPoolKeeper

	// required keepers
	AccountKeeper     *authkeeper.AccountKeeper
	BankKeeper        types.BankKeeper
	OracleKeeper      types.OracleKeeper
	VMKeeper          VMKeeper
	IBCKeeper         *ibckeeper.Keeper
	TransferKeeper    types.TransferKeeper
	NftTransferKeeper types.NftTransferKeeper
	OPChildKeeper     types.OPChildKeeper

	config  *config.IndexerConfig
	homeDir string

	feeCollector string

	schemaBuilder *collections.SchemaBuilder
	schema        *collections.Schema

	ac address.Codec
	vc address.Codec

	db     dbm.DB
	sealed bool

	chainId    string
	crontab    *Crontab
	submodules map[string]Submodule
}

// NewKeeper creates a new indexer Keeper instance
// TODO: remove unncessary arguments
func NewKeeper(
	cdc codec.Codec,
	accountKeeper *authkeeper.AccountKeeper,
	bankKeeper types.BankKeeper,
	oracleKeeper types.OracleKeeper,
	distrKeeper types.DistributionKeeper, // can be nil, if staking not used
	stakingKeeper types.StakingKeeper, // can be nil, if staking not used
	rewardKeeper types.RewardKeeper, // can be nil, if staking not used
	communityPoolKeeper types.CommunityPoolKeeper, // can be nil, if staking not used
	vmKeeper VMKeeper,
	IbcKeeper *ibckeeper.Keeper, // can be nil, if ibc not used
	TransferKeeper types.TransferKeeper, // can be nil, if transfer not used
	NftTransferKeeper types.NftTransferKeeper,
	OPChildKeeper types.OPChildKeeper,
	feeCollector string,
	homeDir string,
	config *config.IndexerConfig,
	ac, vc address.Codec,
	chainId string,
) *Keeper {

	k := &Keeper{
		cdc:                 cdc,
		AccountKeeper:       accountKeeper,
		BankKeeper:          bankKeeper,
		OracleKeeper:        oracleKeeper,
		DistrKeeper:         distrKeeper,
		StakingKeeper:       stakingKeeper,
		RewardKeeper:        rewardKeeper,
		CommunityPoolKeeper: communityPoolKeeper,
		VMKeeper:            vmKeeper,
		IBCKeeper:           IbcKeeper,
		TransferKeeper:      TransferKeeper,
		NftTransferKeeper:   NftTransferKeeper,
		OPChildKeeper:       OPChildKeeper,
		feeCollector:        feeCollector,
		homeDir:             homeDir,
		config:              config,
		schema:              nil,
		ac:                  ac,
		vc:                  vc,
		chainId:             chainId,
		sealed:              false,
	}

	k.crontab = NewCrontab(config, k)
	k.submodules = make(map[string]Submodule)

	sb := collections.NewSchemaBuilderFromAccessor(
		func(ctx context.Context) corestoretypes.KVStore {
			return k.db
		})
	k.schemaBuilder = sb

	return k
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types.ModuleName)
}

func (k *Keeper) Seal() error {
	if k.IsSealed() {
		return errors.New("keeper is already sealed")
	}

	db, err := store.OpenDB(path.Join(k.homeDir, DataDir), StoreName, k.config.BackendConfig)
	if err != nil {
		panic(err)
	}

	schema, err := k.schemaBuilder.Build()
	if err != nil {
		return err
	}

	k.db = db
	k.schema = &schema

	//k.store = cachekv.NewStore(dbadapter.Store{DB: db}) // legacy
	k.store = store.NewStore(dbadapter.Store{DB: db}, 10000)
	k.sealed = true

	return nil
}

func (k Keeper) IsSealed() bool {
	return k.sealed
}

func (k Keeper) GetConfig() *config.IndexerConfig {
	return k.config
}

func (k Keeper) GetStore() *storetypes.CacheKVStore {
	return &k.store
}

func (k *Keeper) WriteStore() error {
	if !k.IsSealed() {
		return errors.New("keeper is not sealed")
	}
	k.store.Write()
	return nil
}

func (k Keeper) GetCodec() codec.Codec {
	return k.cdc
}

func (k Keeper) GetChainId() string {
	return k.chainId
}

func (k Keeper) GetNewAddresses(ctx context.Context, from, to int64) ([]sdk.AccAddress, error) {

	return nil, nil
}

func (k Keeper) GetCrontab() *Crontab {
	return k.crontab
}

func (k Keeper) GetSubmodules() map[string]Submodule {
	return k.submodules
}

func (k Keeper) GetAddressCodec() address.Codec {
	return k.ac
}

func (k Keeper) GetValidatorAddressCodec() address.Codec {
	return k.vc
}

func (k Keeper) GetSchemaBilder() *collections.SchemaBuilder {
	return k.schemaBuilder
}
