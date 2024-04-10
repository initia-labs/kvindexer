package nft

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	cosmoserr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/kvindexer/config"
	"github.com/initia-labs/kvindexer/module/keeper"
	"github.com/initia-labs/kvindexer/submodule/nft/types"
)

func processEvents(k *keeper.Keeper, ctx context.Context, cfg config.SubmoduleConfig, events []types.EventWithAttributeMap) error {
	var fn func(k *keeper.Keeper, ctx context.Context, cfg config.SubmoduleConfig, event types.EventWithAttributeMap) error
	for _, event := range events {
		switch event.AttributesMap["action"] {
		case "mint":
			fn = handleMintEvent
		case "transfer_nft", "send_nft":
			fn = handlerSendOrTransferEvent
		case "burn":
			fn = handleBurnEvent
		default:
			continue
		}
		if err := fn(k, ctx, cfg, event); err != nil {
			k.Logger(ctx).Error("failed to handle nft-related event", "error", err.Error())
			return cosmoserr.Wrap(err, "failed to handle nft-related event")
		}
	}
	return nil
}

func handleMintEvent(k *keeper.Keeper, ctx context.Context, cfg config.SubmoduleConfig, event types.EventWithAttributeMap) error {
	k.Logger(ctx).Debug("minted", "event", event)

	data := types.MintEvent{}
	if err := data.Parse(event); err != nil {
		return cosmoserr.Wrap(err, "failed to parse mint event")
	}

	var collection *types.IndexedCollection
	_, err := collectionMap.Get(ctx, data.ContractAddress)
	if err != nil {
		if !cosmoserr.IsOf(err, collections.ErrNotFound) {
			return cosmoserr.Wrap(err, "failed to check collection existence")
		}
		// if not found, it means this is the first minting of the collection, so we need to set into collectionMap
		collection, err = getIndexedCollectionFromVMStore(k, ctx, data.ContractAddress)
		if err != nil {
			return cosmoserr.Wrap(err, "failed to get collection contract info")
		}

		err = collectionMap.Set(ctx, data.ContractAddress, *collection)
		if err != nil {
			return cosmoserr.Wrap(err, "failed to set collection")
		}
	}

	err = applyCollectionOwnerMap(k, ctx, data.ContractAddress, data.Minter, true)
	if err != nil {
		return cosmoserr.Wrap(err, "failed to insert collection into collectionOwnersMap")
	}

	token, err := getIndexedNftFromVMStore(k, ctx, data.ContractAddress, data.TokenId, &data.Minter)
	if err != nil {
		return cosmoserr.Wrap(err, "failed to get token info")
	}
	token.CollectionName = collection.Collection.Name

	err = tokenMap.Set(ctx, collections.Join(data.ContractAddress, data.TokenId), *token)
	if err != nil {
		return cosmoserr.Wrap(err, "failed to set token")
	}

	err = tokenOwnerMap.Set(ctx, collections.Join3(data.Minter, data.ContractAddress, data.TokenId), true)
	if err != nil {
		k.Logger(ctx).Error("failed to insert into tokenOwnerSet", "minter", data.Minter, "collection-addr", data.ContractAddress, "token-id", token.Nft.TokenId, "error", err)
		return cosmoserr.Wrap(err, "failed to insert into tokenOwnerSet")
	}

	k.Logger(ctx).Info("nft minted", "collection", collection, "token", token)
	return nil
}

func handlerSendOrTransferEvent(k *keeper.Keeper, ctx context.Context, cfg config.SubmoduleConfig, event types.EventWithAttributeMap) (err error) {
	k.Logger(ctx).Info("sent/transferred", "event", event)
	data := types.TransferOrSendEvent{}
	if err := data.Parse(event); err != nil {
		return cosmoserr.Wrap(err, "failed to parse send/transfer event")
	}

	tpk := collections.Join[sdk.AccAddress, string](data.ContractAddress, data.TokenId)

	token, err := tokenMap.Get(ctx, tpk)
	if err != nil {
		k.Logger(ctx).Debug("failed to get nft from prev owner and object addres", "collection-addr", data.ContractAddress, "token-id", data.TokenId, "prevOwner", data.Sender, "error", err)
		return cosmoserr.Wrap(err, "failed to get nft from tokenMap")
	}
	token.OwnerAddr = data.Recipient.String()

	if err = tokenMap.Set(ctx, tpk, token); err != nil {
		return errors.New("failed to delete nft from sender's collection")
	}

	err = applyCollectionOwnerMap(k, ctx, tpk.K1(), data.Sender, false)
	if err != nil {
		return errors.New("failed to decrease collection count from prev owner")

	}
	err = applyCollectionOwnerMap(k, ctx, tpk.K1(), data.Recipient, true)
	if err != nil {
		return errors.New("failed to increase collection count from new owner")
	}

	err = tokenOwnerMap.Remove(ctx, collections.Join3(data.Sender, tpk.K1(), tpk.K2()))
	if err != nil {
		k.Logger(ctx).Error("failed to remove from tokenOwnerSet", "to", data.Recipient, "collection-addr", tpk.K1(), "token-id", tpk.K2(), "error", err)
		return errors.New("failed to insert token into tokenOwnerSet")
	}
	err = tokenOwnerMap.Set(ctx, collections.Join3(data.Recipient, tpk.K1(), tpk.K2()), true)
	if err != nil {
		k.Logger(ctx).Error("failed to insert into tokenOwnerSet", "to", data.Recipient, "collection-addr", tpk.K1(), "token-id", tpk.K2(), "error", err)
		return errors.New("failed to insert token into tokenOwnerSet")
	}

	k.Logger(ctx).Info("nft sent/transferred", "objectKey", tpk, "token", token, "prevOwner", data.Sender, "newOwner", data.Recipient)
	return nil
}

func handleBurnEvent(k *keeper.Keeper, ctx context.Context, cfg config.SubmoduleConfig, event types.EventWithAttributeMap) error {
	k.Logger(ctx).Info("burnt", "event", event)
	cdc := k.GetAddressCodec()

	data := types.BurnEvent{}
	if err := data.Parse(event); err != nil {
		return cosmoserr.Wrap(err, "failed to parse burn event")
	}

	// remove from tokensOwnersMap
	tpk := collections.Join[sdk.AccAddress, string](data.ContractAddress, data.TokenId)
	token, err := tokenMap.Get(ctx, tpk)
	if err != nil {
		return cosmoserr.Wrap(err, "failed to get nft from tokenMap")
	}

	err = tokenMap.Remove(ctx, tpk)
	if err != nil {
		return cosmoserr.Wrap(err, "failed to delete nft from tokenMap")
	}

	ownerAddr, _ := getVMAddress(cdc, token.OwnerAddr)
	ownerSdkAddr := getCosmosAddress(ownerAddr)

	err = tokenOwnerMap.Set(ctx, collections.Join3(ownerSdkAddr, tpk.K1(), tpk.K2()), true)
	if err != nil {
		k.Logger(ctx).Error("failed to remove from tokenOwnerSet", "owner", ownerSdkAddr, "collection-addr", tpk.K1(), "token-id", tpk.K2(), "error", err)
		return cosmoserr.Wrap(err, "failed to insert token into tokenOwnerSet")
	}

	err = applyCollectionOwnerMap(k, ctx, data.ContractAddress, ownerSdkAddr, false)
	if err != nil {
		return err // just return err, no wrap
	}

	k.Logger(ctx).Debug("nft burnt", "event", data)

	return nil
}

func applyCollectionOwnerMap(_ *keeper.Keeper, ctx context.Context, collectionAddr, ownerAddr sdk.AccAddress, isIncrease bool) error {
	count, err := collectionOwnerMap.Get(ctx, collections.Join(ownerAddr, collectionAddr))
	if err != nil {
		if !isIncrease || (isIncrease && !cosmoserr.IsOf(err, collections.ErrNotFound)) {
			return cosmoserr.Wrap(err, "failed to get collection count from collectionOwnersMap")
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
		return cosmoserr.Wrap(err, "failed to update collection count in collectionOwnersMap")
	}
	return nil
}
