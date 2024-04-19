package keeper

import (
	"context"
	"errors"
	"path"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestoretypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/store/dbadapter"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/initia-labs/kvindexer/config"
	"github.com/initia-labs/kvindexer/store"
	"github.com/initia-labs/kvindexer/x/kvindexer/types"
)

const StoreName = "indexer"
const DataDir = "data"

type Keeper struct {
	cdc   codec.Codec
	store *store.CacheStore

	vmType string

	config  *config.IndexerConfig
	homeDir string

	schemaBuilder *collections.SchemaBuilder
	schema        *collections.Schema

	ac address.Codec
	vc address.Codec

	db     dbm.DB
	sealed bool

	submodules map[string]types.Submodule
}

// Close closes indexer goleveldb
func (k Keeper) Close() error {
	if k.db != nil {
		return k.db.Close()
	}

	return nil
}

// NewKeeper creates a new indexer Keeper instance
// TODO: remove unncessary arguments
func NewKeeper(
	cdc codec.Codec,
	vmType string,
	homeDir string,
	config *config.IndexerConfig,
	ac, vc address.Codec,
) *Keeper {

	k := &Keeper{
		cdc:     cdc,
		vmType:  vmType,
		homeDir: homeDir,
		config:  config,
		schema:  nil,
		ac:      ac,
		vc:      vc,
		sealed:  false,
	}

	k.submodules = make(map[string]types.Submodule)

	sb := collections.NewSchemaBuilderFromAccessor(
		func(ctx context.Context) corestoretypes.KVStore {
			return k.store
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

	k.store = store.NewCacheStore(dbadapter.Store{DB: db}, k.config.CacheSize)
	k.sealed = true

	return nil
}

func (k Keeper) IsSealed() bool {
	return k.sealed
}

func (k Keeper) GetSchemaBuilder() *collections.SchemaBuilder {
	return k.schemaBuilder
}

func (k Keeper) GetConfig() *config.IndexerConfig {
	return k.config
}

func (k Keeper) GetStore() *store.CacheStore {
	return k.store
}

func (k Keeper) GetCodec() codec.Codec {
	return k.cdc
}

func (k Keeper) GetSubmodules() map[string]types.Submodule {
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
