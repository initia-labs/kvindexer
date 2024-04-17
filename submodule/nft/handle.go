package nft

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"

	"cosmossdk.io/collections"
	cosmoserr "cosmossdk.io/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/kvindexer/module/keeper"
	"github.com/initia-labs/kvindexer/submodule/nft/types"
	"github.com/initia-labs/kvindexer/submodule/pair"
)

func processEvents(k *keeper.Keeper, ctx context.Context, events []types.EventWithAttributeMap) error {
	var fn func(k *keeper.Keeper, ctx context.Context, event types.EventWithAttributeMap) error
	for _, event := range events {
		if event.Type == "wasm" {
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

	data := types.MintEvent{}
	if err := data.Parse(event); err != nil {
		// may be not nft mint event
		return nil
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

func handlerSendOrTransferEvent(k *keeper.Keeper, ctx context.Context, event types.EventWithAttributeMap) (err error) {
	k.Logger(ctx).Info("sent/transferred", "event", event)
	data := types.TransferOrSendEvent{}
	if err := data.Parse(event); err != nil {
		// may be not nft send/transfer event
		return nil
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

func handleBurnEvent(k *keeper.Keeper, ctx context.Context, event types.EventWithAttributeMap) error {
	k.Logger(ctx).Info("burnt", "event", event)
	cdc := k.GetAddressCodec()

	data := types.BurnEvent{}
	if err := data.Parse(event); err != nil {
		// may be not nft burn event
		return nil
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

	k.Logger(ctx).Info("nft burnt", "event", data)

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

func handleWriteAcknowledgementEvent(k *keeper.Keeper, ctx context.Context, attrs []abci.EventAttribute) (err error) {
	k.Logger(ctx).Debug("write-ack", "attrs", attrs)
	for _, attr := range attrs {
		if attr.Key != "packet_data" {
			continue
		}

		data := types.WriteAckForNftEvent{}
		if err = json.Unmarshal([]byte(attr.Value), &data); err != nil {
			// may be not target
			return nil
		}

		cdb, err := base64.StdEncoding.DecodeString(data.ClassData)
		if err != nil {
			return cosmoserr.Wrap(err, "failed to decode class data")
		}
		classData := types.NftClassData{}
		if err = json.Unmarshal(cdb, &classData); err != nil {
			return cosmoserr.Wrap(err, "failed to unmarshal class data")
		}

		_, err = pair.GetPair(ctx, false, data.ClassId)
		if err == nil {
			return nil // already exists
		}
		if !cosmoserr.IsOf(err, collections.ErrNotFound) {
			return cosmoserr.Wrap(err, "failed to check class existence")
		}

		err = pair.SetPair(ctx, false, false, data.ClassId, classData.Description.Value)
		if err != nil {
			return cosmoserr.Wrap(err, "failed to set class")
		}

		k.Logger(ctx).Info("nft class added", "classId", data.ClassId, "description", classData.Description.Value)
	}
	return nil
}
