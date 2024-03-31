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
const tokenAddressIndexPrefix = 0x31
const tokenOwnerIndexPrefix = 0x32

const collectionMapName = "collections"
const collectionOwnersMapName = "collection_owners"

const tokenMapName = "tokens"
const tokenAddressIndexName = "token_addr"
const tokenOwnerIndexName = "owner_addr"

var (
	prefixCollection        = keeper.NewPrefix(submoduleName, collectionsPrefix)
	prefixCollectionOwners  = keeper.NewPrefix(submoduleName, collectionOwnersPrefix)
	prefixTokens            = keeper.NewPrefix(submoduleName, tokenPrefix)
	prefixTokenAddressIndex = keeper.NewPrefix(submoduleName, tokenAddressIndexPrefix)
	prefixTokenOwnerIndex   = keeper.NewPrefix(submoduleName, tokenOwnerIndexPrefix)
)

//
// maps
//

// key: collectiona-address, value: indexed collection
var collectionMap *collections.Map[sdk.AccAddress, types.IndexedCollection]

// key: pair[owner-address, collection-address], value: holding token count
var collectionOwnerMap *collections.Map[collections.Pair[sdk.AccAddress, sdk.AccAddress], uint64]

// key: pair[collection-address, token_id], value: indexed token
var tokenMap *collections.IndexedMap[collections.Pair[sdk.AccAddress, string], types.IndexedToken, TokenIndex]

//
// Indices - vm specific
//

type CollectionIndex struct {
	// ref: owner-address, pk: collection-address, value: collection
	OwnerAddress *indexes.Multi[ /*ref*/ sdk.AccAddress /*pk*/, sdk.AccAddress /*val*/, types.IndexedCollection]
}

type TokenIndex struct {
	// ref: token-address, pk: pair[collection-address, token-id], value: indexed token
	TokenAddress *indexes.Unique[ /*ref*/ sdk.AccAddress /*pk*/, collections.Pair[sdk.AccAddress, string] /*val*/, types.IndexedToken]
	// ref: owner-address, pk: pair[collection-address, token-id], value: indexed token
	OwnerAddress *indexes.Multi[ /*ref*/ sdk.AccAddress /*pk*/, collections.Pair[sdk.AccAddress, string] /*val*/, types.IndexedToken]
}

func (i TokenIndex) IndexesList() []collections.Index[collections.Pair[sdk.AccAddress, string], types.IndexedToken] {
	return []collections.Index[collections.Pair[sdk.AccAddress, string], types.IndexedToken]{
		i.TokenAddress,
		i.OwnerAddress,
	}
}

func newTokensIndex(k *keeper.Keeper) TokenIndex {
	cdc := k.GetAddressCodec()
	return TokenIndex{
		TokenAddress: indexes.NewUnique(
			k.GetSchemaBilder(),     // schema builder
			prefixTokenAddressIndex, // prefix
			tokenAddressIndexName,   // name
			sdk.AccAddressKey,       // refCodec
			collections.PairKeyCodec(sdk.AccAddressKey, collections.StringKey), // pkCodec
			func(k collections.Pair[sdk.AccAddress, string], v types.IndexedToken) (sdk.AccAddress, error) {
				vmAddr, err := getVMAddress(cdc, v.ObjectAddr)
				if err != nil {
					return sdk.AccAddress{}, err
				}
				return getCosmosAddress(vmAddr), nil
			}, // getRefKeyFunc
		),
		OwnerAddress: indexes.NewMulti(
			k.GetSchemaBilder(),   // schema builder
			prefixTokenOwnerIndex, // prefix
			tokenOwnerIndexName,   // name
			sdk.AccAddressKey,     // refCodec
			collections.PairKeyCodec(sdk.AccAddressKey, collections.StringKey), // pkCodec
			func(k collections.Pair[sdk.AccAddress, string], v types.IndexedToken) (sdk.AccAddress, error) {
				return sdk.AccAddressFromBech32(v.OwnerAddr)
			}, // getRefKeyFunc
		),
	}
}

func addStorages(k *keeper.Keeper, _ context.Context, _ config.SubmoduleConfig) (err error) {
	cdc := k.GetCodec()

	if collectionMap, err = keeper.AddMap(k, prefixCollection, collectionMapName, sdk.AccAddressKey, codec.CollValue[types.IndexedCollection](cdc)); err != nil {
		return err
	}

	if collectionOwnerMap, err = keeper.AddMap(k, prefixCollectionOwners, collectionOwnersMapName, collections.PairKeyCodec(sdk.AccAddressKey, sdk.AccAddressKey), collections.Uint64Value); err != nil {
		return err
	}

	if tokenMap, err = keeper.AddIndexedMap(k, prefixTokens, tokenMapName, collections.PairKeyCodec(sdk.AccAddressKey, collections.StringKey), codec.CollValue[types.IndexedToken](cdc), newTokensIndex(k)); err != nil {
		return err
	}

	return nil
}
