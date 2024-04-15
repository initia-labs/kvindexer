package nft

import (
	"context"
	"encoding/json"
	"errors"

	"cosmossdk.io/collections"
	cosmoserr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	movetypes "github.com/initia-labs/initia/x/move/types"
	"github.com/initia-labs/kvindexer/module/keeper"
	"github.com/initia-labs/kvindexer/submodule/nft/types"
)

func processEvents(k *keeper.Keeper, ctx context.Context, events []types.EventWithAttributeMap) error {
	var fn func(k *keeper.Keeper, ctx context.Context, event types.EventWithAttributeMap) error
	for _, event := range events {
		switch event.AttributesMap["type_tag"] {
		case "0x1::collection::MintEvent":
			fn = handleMintEvent
		case "0x1::object::TransferEvent":
			fn = handlerTransferEvent
		case "0x1::nft::MutationEvent", "0x1::collection::MutationEvent":
			fn = handleMutateEvent
		case "0x1::collection::BurnEvent":
			fn = handleBurnEvent
		default:
			continue
		}
		if err := fn(k, ctx, event); err != nil {
			k.Logger(ctx).Error("failed to handle nft-related event", "error", err.Error())
			return cosmoserr.Wrap(err, "failed to handle nft-related event")
		}
	}
	return nil
}

func handleMintEvent(k *keeper.Keeper, ctx context.Context, event types.EventWithAttributeMap) error {
	k.Logger(ctx).Debug("minted", "event", event)

	data := types.NftMintAndBurnEventData{}
	if err := json.Unmarshal([]byte(event.AttributesMap["data"]), &data); err != nil {
		return errors.New("failed to unmarshal mint event")
	}

	collectionAddr, err := getVMAddress(k.GetAddressCodec(), data.Collection)
	if err != nil {
		return errors.New("failed to parse collection address")
	}

	collection, err := getIndexedCollectionFromVMStore(k, ctx, collectionAddr)
	if err != nil {
		return errors.New("failed to get minted collection info")
	}

	tokenAddr, err := getVMAddress(k.GetAddressCodec(), data.Nft)
	if err != nil {
		return errors.New("failed to parse nft address")
	}

	creatorAddr, err := getVMAddress(k.GetAddressCodec(), collection.Collection.Creator)
	if err != nil {
		return errors.New("failed to parse creator address")
	}

	creatorSdkAddr := getCosmosAddress(creatorAddr)
	collectionSdkAddr := getCosmosAddress(collectionAddr)
	tokenSdkAddr := getCosmosAddress(tokenAddr)

	token, err := getIndexedTokenFromVMStore(k, ctx, tokenAddr, &collectionAddr)
	if err != nil {
		return errors.New("failed to get minted nft info")
	}
	token.CollectionName = collection.Collection.Name
	token.OwnerAddr = creatorSdkAddr.String()

	_, err = collectionMap.Get(ctx, collectionSdkAddr)
	if err != nil {
		if !cosmoserr.IsOf(err, collections.ErrNotFound) {
			return errors.New("")
		}
		err = collectionMap.Set(ctx, collectionSdkAddr, *collection)
		if err != nil {
			return errors.New("failed to insert collection into collectionMap")
		}
	}

	err = applyCollectionOwnerMap(k, ctx, collectionSdkAddr, creatorSdkAddr, true)
	if err != nil {
		return errors.New("failed to insert collection into collectionOwnersMap")
	}

	err = tokenMap.Set(ctx, collections.Join(collectionSdkAddr, token.Nft.TokenId), *token)
	if err != nil {
		k.Logger(ctx).Error("failed to insert token into tokenMap", "collection-addr", collectionSdkAddr, "token-id", token.Nft.TokenId, "error", err, "token", token)
		return errors.New("failed to insert token into tokenMap")
	}
	err = tokenOwnerMap.Set(ctx, collections.Join3(creatorSdkAddr, collectionSdkAddr, token.Nft.TokenId), true)
	if err != nil {
		k.Logger(ctx).Error("failed to insert into tokenOwnerSet", "owner", creatorSdkAddr, "collection-addr", collectionSdkAddr, "token-id", token.Nft.TokenId, "error", err)
		return errors.New("failed to insert into tokenOwnerSet")
	}

	k.Logger(ctx).Info("nft minted", "collection", collection, "nft", token, "collection-sdk-addr", collectionSdkAddr, "nft-sdk-addr", tokenSdkAddr, "creator-sdk-addr", creatorSdkAddr)
	return nil
}

func handlerTransferEvent(k *keeper.Keeper, ctx context.Context, event types.EventWithAttributeMap) error {
	k.Logger(ctx).Info("transferred", "event", event)

	data := types.NftTransferEventData{}
	if err := json.Unmarshal([]byte(event.AttributesMap["data"]), &data); err != nil {
		return errors.New("failed to unmarshal transfer event")
	}

	objectAddr, err := getVMAddress(k.GetAddressCodec(), data.Object)
	if err != nil {
		return errors.New("failed to parse object address")
	}
	objectSdkAddr := getCosmosAddress(objectAddr)

	fromAddr, err := movetypes.AccAddressFromString(k.GetAddressCodec(), data.From)
	if err != nil {
		return errors.New("failed to parse prev owner address")
	}
	fromSdkAddr := getCosmosAddress(fromAddr)

	toAddr, err := getVMAddress(k.GetAddressCodec(), data.To)
	if err != nil {
		return errors.New("failed to parse new owner address")
	}
	toSdkAddr := getCosmosAddress(toAddr)

	tpk, err := tokenMap.Indexes.TokenAddress.MatchExact(ctx, objectSdkAddr)
	if err != nil {
		return errors.New("token's object address not found")
	}

	token, err := tokenMap.Get(ctx, tpk)
	if err != nil {
		// NOT all transferEvent means the nft is transferred. it's all object transfer event. so it's okay to ignore NotFound error
		if cosmoserr.IsOf(err, collections.ErrNotFound) {
			k.Logger(ctx).Debug("nft not found, maybe not NFT related object transfer", "object", objectSdkAddr.String(), "prevOwner", fromSdkAddr.String())
			return nil
		}
		k.Logger(ctx).Info("failed to get nft from prev owner and object addres", "err", err, "object", objectSdkAddr.String(), "prev", fromSdkAddr.String())

		return err
	}
	token.OwnerAddr = toSdkAddr.String()

	if err = tokenMap.Set(ctx, tpk, token); err != nil {
		return errors.New("failed to delete nft from sender's collection")
	}

	err = applyCollectionOwnerMap(k, ctx, tpk.K1(), fromSdkAddr, false)
	if err != nil {
		return errors.New("failed to decrease collection count from prev owner")

	}
	err = applyCollectionOwnerMap(k, ctx, tpk.K1(), toSdkAddr, true)
	if err != nil {
		return errors.New("failed to increase collection count from new owner")
	}

	err = tokenOwnerMap.Remove(ctx, collections.Join3(fromSdkAddr, tpk.K1(), tpk.K2()))
	if err != nil {
		k.Logger(ctx).Error("failed to remove from tokenOwnerSet", "to", toSdkAddr, "collection-addr", tpk.K1(), "token-id", tpk.K2(), "error", err)
		return errors.New("failed to insert token into tokenOwnerSet")
	}
	err = tokenOwnerMap.Set(ctx, collections.Join3(toSdkAddr, tpk.K1(), tpk.K2()), true)
	if err != nil {
		k.Logger(ctx).Error("failed to insert into tokenOwnerSet", "to", toSdkAddr, "collection-addr", tpk.K1(), "token-id", tpk.K2(), "error", err)
		return errors.New("failed to insert token into tokenOwnerSet")
	}

	k.Logger(ctx).Info("nft transferred", "objectKey", tpk, "token", token, "prevOwner", data.From, "newOwner", data.To)
	return nil
}

func handleMutateEvent(k *keeper.Keeper, ctx context.Context, event types.EventWithAttributeMap) error {
	k.Logger(ctx).Info("mutated", "event", event)
	cdc := k.GetAddressCodec()

	data := types.MutationEventData{}
	if err := json.Unmarshal([]byte(event.AttributesMap["data"]), &data); err != nil {
		return errors.New("failed to unmarshal mutation event")
	}

	switch {
	case data.Nft != "":
		objectAddr, err := getVMAddress(cdc, data.Nft)
		if err != nil {
			return errors.New("failed to parse object address")
		}
		objectSdkAddr := getCosmosAddress(objectAddr)

		nft, err := getIndexedTokenFromVMStore(k, ctx, objectAddr, nil)
		if err != nil {
			return errors.New("failed to get minted nft info")
		}
		k.Logger(ctx).Debug("mutated", "nft", nft)

		// remove the nft from the sender's collection
		tpk, err := tokenMap.Indexes.TokenAddress.MatchExact(ctx, objectSdkAddr)
		//objectKey, err := nftByOwner.TokenAddress.MatchExact(ctx, objectSdkAddr)
		if err != nil {
			return errors.New("object key not found")
		}
		nft.CollectionAddr = tpk.K1().String()

		if err = tokenMap.Set(ctx, tpk, *nft); err != nil {
			return errors.New("failed to update mutated nft")
		}
	case data.Collection != "":
		collectionAddr, err := getVMAddress(cdc, data.Collection)
		if err != nil {
			return errors.New("failed to parse object address")
		}

		collection, err := getIndexedCollectionFromVMStore(k, ctx, collectionAddr)
		if err != nil {
			return errors.New("failed to get mutated collection info")
		}

		err = collectionMap.Set(ctx, getCosmosAddress(collectionAddr), *collection)
		if err != nil {
			return errors.New("failed to update mutated collection")
		}
	}

	k.Logger(ctx).Info("nft mutated", "nft", data.Nft, "collection", data.Collection)

	return nil
}

func handleBurnEvent(k *keeper.Keeper, ctx context.Context, event types.EventWithAttributeMap) error {
	k.Logger(ctx).Info("burnt", "event", event)
	cdc := k.GetAddressCodec()
	burnt := types.NftMintAndBurnEventData{}
	if err := json.Unmarshal([]byte(event.AttributesMap["data"]), &burnt); err != nil {
		return errors.New("failed to unmarshal burnt event")
	}

	objectAddr, err := getVMAddress(cdc, burnt.Nft)
	if err != nil {
		return errors.New("failed to parse object address")
	}
	objectSdkAddr := getCosmosAddress(objectAddr)

	// remove from tokensOwnersMap
	tpk, err := tokenMap.Indexes.TokenAddress.MatchExact(ctx, objectSdkAddr)
	if err != nil {
		return errors.New("token's object address not found")
	}
	token, err := tokenMap.Get(ctx, tpk)
	if err != nil {
		return errors.New("failed to get nft from tokenMap")
	}

	err = tokenMap.Remove(ctx, tpk)
	if err != nil {
		return errors.New("failed to delete nft from tokenMap")
	}

	collectionAddr, _ := getVMAddress(cdc, token.CollectionAddr)
	collectionSdkAddr := getCosmosAddress(collectionAddr)

	ownerAddr, _ := getVMAddress(cdc, token.OwnerAddr)
	ownerSdkAddr := getCosmosAddress(ownerAddr)

	err = applyCollectionOwnerMap(k, ctx, collectionSdkAddr, ownerSdkAddr, false)
	if err != nil {
		return err // just return err, no wrap
	}

	err = tokenOwnerMap.Set(ctx, collections.Join3(ownerSdkAddr, tpk.K1(), tpk.K2()), true)
	if err != nil {
		k.Logger(ctx).Error("failed to remove from tokenOwnerSet", "owner", ownerSdkAddr, "collection-addr", tpk.K1(), "token-id", tpk.K2(), "error", err)
		return errors.New("failed to insert token into tokenOwnerSet")
	}

	k.Logger(ctx).Info("nft burnt", "event", burnt)
	return nil
}

func applyCollectionOwnerMap(_ *keeper.Keeper, ctx context.Context, collectionAddr, ownerAddr sdk.AccAddress, isIncrease bool) error {
	count, err := collectionOwnerMap.Get(ctx, collections.Join(ownerAddr, collectionAddr))
	if err != nil {
		if !isIncrease || (isIncrease && !cosmoserr.IsOf(err, collections.ErrNotFound)) {
			return errors.New("failed to get collection count from collectionOwnersMap")
		}
	}
	if isIncrease {
		count++
	} else {
		count--
	}

	if count == 0 {
		err = collectionOwnerMap.Remove(ctx, collections.Join(ownerAddr, collectionAddr))
	} else {
		err = collectionOwnerMap.Set(ctx, collections.Join(ownerAddr, collectionAddr), count)
	}
	if err != nil {
		return errors.New("failed to update collection count in collectionOwnersMap")
	}
	return nil
}
