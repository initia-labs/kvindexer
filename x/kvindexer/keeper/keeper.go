package keeper

import (
	"context"
	"errors"
	"sync/atomic"

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

type Keeper struct {
	cdc   codec.Codec
	store *store.CacheStore

	vmType string

	config *config.IndexerConfig

	schemaBuilder *collections.SchemaBuilder
	schema        *collections.Schema

	ac address.Codec
	vc address.Codec

	db     dbm.DB
	sealed bool

	submodules []types.Submodule

	pruningRunning *atomic.Bool
}

// Close closes indexer goleveldb
func (k Keeper) Close() error {
	if k.db != nil {
		return k.db.Close()
	}

	return nil
}

// NewKeeper creates a new indexer Keeper instance
// TODO: remove unnecessary arguments
func NewKeeper(
	cdc codec.Codec,
	vmType string,
	db dbm.DB,
	config *config.IndexerConfig,
	ac, vc address.Codec,
) *Keeper {

	k := &Keeper{
		cdc:            cdc,
		vmType:         vmType,
		db:             db,
		config:         config,
		schema:         nil,
		ac:             ac,
		vc:             vc,
		sealed:         false,
		pruningRunning: &atomic.Bool{},
	}

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

	// just mark it as sealed if the indexer is disabled
	if !k.config.IsEnabled() {
		k.sealed = true
		return nil
	}

	schema, err := k.schemaBuilder.Build()
	if err != nil {
		return err
	}

	k.schema = &schema

	k.store = store.NewCacheStore(dbadapter.Store{DB: k.db}, k.config.CacheCapacity)
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

func (k Keeper) GetSubmodules() []types.Submodule {
	return k.submodules
}

func (k Keeper) GetAddressCodec() address.Codec {
	return k.ac
}

func (k Keeper) GetValidatorAddressCodec() address.Codec {
	return k.vc
}
