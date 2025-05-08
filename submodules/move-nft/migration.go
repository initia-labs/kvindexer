package move_nft

import (
	"context"
	"strings"
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

func (sm MoveNftSubmodule) migrateHandler(ctx context.Context) (err error) {
	migrated.Do(func() {

		value, e := sm.migrationInfo.Get(ctx, keyMigrateCollectionName)
		if err != nil {
			if !cosmoserr.IsOf(err, collections.ErrNotFound) {
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
				err = e
				return
			}
			err = sm.migrationInfo.Set(ctx, keyMigrateCollectionName, "v1.0.0")
			if err != nil {
				err = e
				return
			}
		}
	})

	return err
}

// migrateCollectionName_1_0_0 migrates the collection name to lower case and sets it in the collectionNameMap.
func (sm MoveNftSubmodule) migrateCollectionName_1_0_0(ctx context.Context) error {

	// itertate over all collections
	sm.collectionMap.Walk(ctx, nil, func(key sdk.AccAddress, value types.IndexedCollection) (bool, error) {
		pairName, err := sm.getCollectionNameFromPairSubmodule(ctx, value.Collection.Name)
		if err != nil {
			return false, err
		}
		err = sm.applyCollectionNameMap(ctx, pairName, key)
		sm.Logger(ctx).Info("migrating collection name", "original-name", value.Collection.Name, "pair-name", pairName, "address", key.String())
		return err != nil, err
	})

	return nil
}

// applyCollectionNameMap applies the collection name map to the lowercased collection name.
func (sm MoveNftSubmodule) applyCollectionNameMap(ctx context.Context, name string, addr sdk.AccAddress) error {
	// use lowercased name to support case insensitive search
	name = strings.ToLower(name)

	addrs, err := sm.collectionNameMap.Get(ctx, name)
	if err != nil {
		if !cosmoserr.IsOf(err, collections.ErrNotFound) {
			return err
		}
	}
	addrs = appendString(addrs, addr.String())
	err = sm.collectionNameMap.Set(ctx, name, addrs)
	if err != nil {
		return err
	}

	return nil
}

// appendString appends two strings with a comma separator.
func appendString(s1, s2 string) string {
	if s1 == "" {
		return s2
	}
	return s1 + "," + s2
}

func expandString(s []string) (res []string) {
	for _, v := range s {
		res = append(res, strings.Split(v, ",")...)
	}
	return res
}
