package evm_nft

import (
	"context"
	"sync"

	"cosmossdk.io/collections"
	cosmoserr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/kvindexer/nft/types"
	"golang.org/x/mod/semver"
)

const (
	keyMigrateCollectionName = "migrate-collection-name"
)

var migrated sync.Once

func (sm EvmNFTSubmodule) migrateHandler(ctx context.Context) (err error) {
	migrated.Do(func() {

		value, e := sm.migrationInfo.Get(ctx, keyMigrateCollectionName)
		if e != nil {
			if !cosmoserr.IsOf(e, collections.ErrNotFound) {
				err = e
				return
			}
			// if not found, it means migration is needed.
			value = "v0.0.0"
		}
		// if current semver is less than v1.0.0, then migration is needed
		if semver.Compare(value, "v1.0.0") < 0 {
			// do migration
			err = sm.migrateCollectionName_1_0_0(ctx)
			if err != nil {
				return
			}
			err = sm.migrationInfo.Set(ctx, keyMigrateCollectionName, "v1.0.0")
			if err != nil {
				return
			}
		}
	})

	return err
}

// migrateCollectionName_1_0_0 migrates the collection name to lower case and sets it in the collectionNameMap.
func (sm EvmNFTSubmodule) migrateCollectionName_1_0_0(ctx context.Context) error {

	// itertate over all collections
	return sm.collectionMap.Walk(ctx, nil, func(key sdk.AccAddress, value types.IndexedCollection) (bool, error) {
		err := sm.applyCollectionNameMap(ctx, value.Collection.Name, key)
		sm.Logger(ctx).Info("migrating collection name", "original-name", value.Collection.Name, "address", key.String())
		return err != nil, err
	})
}
