package wasm_nft

import (
	"context"
	"regexp"
	"sort"
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

// regexStripNonAlnum is used to strip non-alphanumeric characters from the collection name.
var regexStripNonAlnum = regexp.MustCompile("[^a-zA-Z0-9]+")

var migrated sync.Once

func (sm WasmNFTSubmodule) migrateHandler(ctx context.Context) (err error) {
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
func (sm WasmNFTSubmodule) migrateCollectionName_1_0_0(ctx context.Context) error {

	// itertate over all collections
	return sm.collectionMap.Walk(ctx, nil, func(key sdk.AccAddress, value types.IndexedCollection) (bool, error) {
		err := sm.applyCollectionNameMap(ctx, value.Collection.Name, key)
		sm.Logger(ctx).Info("migrating collection name", "original-name", value.Collection.Name, "address", key.String())
		return err != nil, err
	})
}

// applyCollectionNameMap applies the collection name map to the lowercased collection name.
func (sm WasmNFTSubmodule) applyCollectionNameMap(ctx context.Context, name string, addr sdk.AccAddress) error {
	// use lowercased name to support case insensitive search
	name, _ = sm.getCollectionNameFromPairSubmodule(ctx, name)
	name = strings.ToLower(stripNonAlnum(name))

	addrs, err := sm.collectionNameMap.Get(ctx, name)
	if err != nil {
		if !cosmoserr.IsOf(err, collections.ErrNotFound) {
			return err
		}
	}
	newaddrs := appendString(addrs, addr.String())
	if newaddrs == addrs {
		return nil
	}
	err = sm.collectionNameMap.Set(ctx, name, newaddrs)
	if err != nil {
		return err
	}

	return nil
}

// appendString appends two strings with a comma separator.
func appendString(s1, s2 string) string {
	strs := expandString([]string{s1, s2})

	strmap := make(map[string]bool)
	for _, str := range strs {
		strmap[str] = true
	}

	uniquestrs := make([]string, 0, len(strmap))
	for str := range strmap {
		if str == "" {
			continue
		}
		uniquestrs = append(uniquestrs, str)
	}
	sort.Strings(uniquestrs)
	return strings.Join(uniquestrs, ",")
}

func expandString(s []string) (res []string) {
	for _, v := range s {
		res = append(res, strings.Split(v, ",")...)
	}
	return res
}

func stripNonAlnum(in string) string {
	return regexStripNonAlnum.ReplaceAllString(in, "")
}
