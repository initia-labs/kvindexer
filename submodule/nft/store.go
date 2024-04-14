package nft

import (
	"context"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/kvindexer/config"
	"github.com/initia-labs/kvindexer/module/keeper"
	"github.com/initia-labs/kvindexer/submodule/nft/types"
)

//
// constants and variables
//

const collectionsPrefix = 0x10
const collectionOwnersPrefix = 0x20
const tokenPrefix = 0x30
const tokenOwnerPrefix = 0x40

const collectionMapName = "collections"
const collectionOwnersMapName = "collection_owners"
const tokenMapName = "tokens"
const tokenOwnerSetName = "token_owner"

var (
	prefixCollection       = keeper.NewPrefix(submoduleName, collectionsPrefix)
	prefixCollectionOwners = keeper.NewPrefix(submoduleName, collectionOwnersPrefix)
	prefixTokens           = keeper.NewPrefix(submoduleName, tokenPrefix)
	prefixTokenOwner       = keeper.NewPrefix(submoduleName, tokenOwnerPrefix)
)

//
// maps
//

// key: collectiona-address, value: indexed collection
var collectionMap *collections.Map[sdk.AccAddress, types.IndexedCollection]

// key: pair[owner-address, collection-address], value: holding token count
var collectionOwnerMap *collections.Map[collections.Pair[sdk.AccAddress, sdk.AccAddress], uint64]

// key: pair[collection-address, token_id], value: indexed token
var tokenMap *collections.Map[collections.Pair[sdk.AccAddress, string], types.IndexedToken]

// key: triple[owner-addr, collection-address, token-id], value: none
var tokenOwnerMap *collections.Map[collections.Triple[sdk.AccAddress, sdk.AccAddress, string], bool]

//
// Indices - vm specific
//

type CollectionIndex struct {
	// ref: owner-address, pk: collection-address, value: collection
	OwnerAddress *indexes.Multi[ /*ref*/ sdk.AccAddress /*pk*/, sdk.AccAddress /*val*/, types.IndexedCollection]
}

func addStorages(k *keeper.Keeper, _ context.Context, _ config.SubmoduleConfig) (err error) {
	cdc := k.GetCodec()

	if collectionMap, err = keeper.AddMap(k, prefixCollection, collectionMapName, sdk.AccAddressKey, codec.CollValue[types.IndexedCollection](cdc)); err != nil {
		return err
	}

	if collectionOwnerMap, err = keeper.AddMap(k, prefixCollectionOwners, collectionOwnersMapName, collections.PairKeyCodec(sdk.AccAddressKey, sdk.AccAddressKey), collections.Uint64Value); err != nil {
		return err
	}

	if tokenMap, err = keeper.AddMap(k, prefixTokens, tokenMapName, collections.PairKeyCodec(sdk.AccAddressKey, collections.StringKey), codec.CollValue[types.IndexedToken](cdc)); err != nil {
		return err
	}

	if tokenOwnerMap, err = keeper.AddMap(k, prefixTokenOwner, tokenOwnerSetName, collections.TripleKeyCodec(sdk.AccAddressKey, sdk.AccAddressKey, collections.StringKey), collections.BoolValue); err != nil {
		return err
	}

	return nil
}
