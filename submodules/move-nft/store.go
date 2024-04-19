package move_nft

import (
	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/initia-labs/kvindexer/collection"
	nfttypes "github.com/initia-labs/kvindexer/internal/nft/types"
	"github.com/initia-labs/kvindexer/submodules/move-nft/types"
)

//
// Indices - vm specific
//

type CollectionIndex struct {
	// ref: owner-address, pk: collection-address, value: collection
	OwnerAddress *indexes.Multi[ /*ref*/ sdk.AccAddress /*pk*/, sdk.AccAddress /*val*/, nfttypes.IndexedCollection]
}

type TokenIndex struct {
	// ref: token-address, pk: pair[collection-address, token-id], value: indexed token
	TokenAddress *indexes.Unique[ /*ref*/ sdk.AccAddress /*pk*/, collections.Pair[sdk.AccAddress, string] /*val*/, nfttypes.IndexedToken]
}

func (i TokenIndex) IndexesList() []collections.Index[collections.Pair[sdk.AccAddress, string], nfttypes.IndexedToken] {
	return []collections.Index[collections.Pair[sdk.AccAddress, string], nfttypes.IndexedToken]{
		i.TokenAddress,
		//i.OwnerAddress,
	}
}

func newTokensIndex(ac address.Codec, k collection.IndexerKeeper) TokenIndex {
	tokensPrefix := collection.NewPrefix(types.SubmoduleName, types.TokenAddressIndexPrefix)
	return TokenIndex{
		TokenAddress: indexes.NewUnique(
			k.GetSchemaBuilder(),  // schema builder
			tokensPrefix,          // prefix
			"token_address_index", // name
			sdk.AccAddressKey,     // refCodec
			collections.PairKeyCodec(sdk.AccAddressKey, collections.StringKey), // pkCodec
			func(k collections.Pair[sdk.AccAddress, string], v nfttypes.IndexedToken) (sdk.AccAddress, error) {
				vmAddr, err := getVMAddress(ac, v.ObjectAddr)
				if err != nil {
					return sdk.AccAddress{}, err
				}
				return getCosmosAddress(vmAddr), nil
			}, // getRefKeyFunc
		),
	}
}
