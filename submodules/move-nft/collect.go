package move_nft

import (
	"context"
	"encoding/json"
	"errors"

	"cosmossdk.io/collections"
	cosmoserr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/cometbft/cometbft/abci/types"

	movetypes "github.com/initia-labs/initia/x/move/types"

	"github.com/initia-labs/kvindexer/submodules/move-nft/types"
)

func (sm MoveNftSubmodule) finalizeBlock(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	sm.Logger(ctx).Debug("finalizeBlock", "submodule", types.SubmoduleName, "txs", len(req.Txs), "height", req.Height)

	for _, txResult := range res.TxResults {
		events := filterAndParseEvent(eventType, txResult.Events)
		err := sm.processEvents(ctx, events)
		if err != nil {
			sm.Logger(ctx).Debug("processEvents", "error", err)
		}
	}

	return nil
}

func (sm MoveNftSubmodule) processEvents(ctx context.Context, events []types.EventWithAttributeMap) error {
	var fn func(ctx context.Context, event types.EventWithAttributeMap) error
	for _, event := range events {
		switch event.AttributesMap["type_tag"] {
		case "0x1::collection::MintEvent":
			fn = sm.handleMintEvent
		case "0x1::object::TransferEvent":
			fn = sm.handlerTransferEvent
		case "0x1::nft::MutationEvent", "0x1::collection::MutationEvent":
			fn = sm.handleMutateEvent
		case "0x1::collection::BurnEvent":
			fn = sm.handleBurnEvent
		default:
			continue
		}
		if err := fn(ctx, event); err != nil {
			sm.Logger(ctx).Info("failed to handle nft-related event", "error", err.Error())
			// don't return here because we want to process all events
		}
	}
	return nil
}

func (sm MoveNftSubmodule) handleMintEvent(ctx context.Context, event types.EventWithAttributeMap) error {
	sm.Logger(ctx).Debug("minted", "event", event)

	data := types.NftMintAndBurnEventData{}
	if err := json.Unmarshal([]byte(event.AttributesMap["data"]), &data); err != nil {
		return errors.New("failed to unmarshal mint event")
	}

	collectionAddr, err := getVMAddress(sm.ac, data.Collection)
	if err != nil {
		return errors.New("failed to parse collection address")
	}

	collection, err := sm.getIndexedCollectionFromVMStore(ctx, collectionAddr)
	if err != nil {
		return errors.New("failed to get minted collection info")
	}

	tokenAddr, err := getVMAddress(sm.ac, data.Nft)
	if err != nil {
		return errors.New("failed to parse nft address")
	}

	creatorAddr, err := getVMAddress(sm.ac, collection.Collection.Creator)
	if err != nil {
		return errors.New("failed to parse creator address")
	}

	creatorSdkAddr := getCosmosAddress(creatorAddr)
	collectionSdkAddr := getCosmosAddress(collectionAddr)
	tokenSdkAddr := getCosmosAddress(tokenAddr)

	token, err := sm.getIndexedTokenFromVMStore(ctx, tokenAddr, &collectionAddr)
	if err != nil {
		return errors.New("failed to get minted nft info")
	}
	token.CollectionName = collection.Collection.Name
	token.OwnerAddr = creatorSdkAddr.String()

	_, err = sm.collectionMap.Get(ctx, collectionSdkAddr)
	if err != nil {
		if !cosmoserr.IsOf(err, collections.ErrNotFound) {
			return errors.New("")
		}
		err = sm.collectionMap.Set(ctx, collectionSdkAddr, *collection)
		if err != nil {
			return errors.New("failed to insert collection into collectionMap")
		}
	}

	err = sm.applyCollectionOwnerMap(ctx, collectionSdkAddr, creatorSdkAddr, true)
	if err != nil {
		return errors.New("failed to insert collection into collectionOwnersMap")
	}

	err = sm.tokenMap.Set(ctx, collections.Join(collectionSdkAddr, token.Nft.TokenId), *token)
	if err != nil {
		sm.Logger(ctx).Error("failed to insert token into tokenMap", "collection-addr", collectionSdkAddr, "token-id", token.Nft.TokenId, "error", err, "token", token)
		return errors.New("failed to insert token into tokenMap")
	}
	err = sm.tokenOwnerMap.Set(ctx, collections.Join3(creatorSdkAddr, collectionSdkAddr, token.Nft.TokenId), true)
	if err != nil {
		sm.Logger(ctx).Error("failed to insert into tokenOwnerSet", "owner", creatorSdkAddr, "collection-addr", collectionSdkAddr, "token-id", token.Nft.TokenId, "error", err)
		return errors.New("failed to insert into tokenOwnerSet")
	}

	sm.Logger(ctx).Info("nft minted", "collection", collection, "nft", token, "collection-sdk-addr", collectionSdkAddr, "nft-sdk-addr", tokenSdkAddr, "creator-sdk-addr", creatorSdkAddr)
	return nil
}

func (sm MoveNftSubmodule) handlerTransferEvent(ctx context.Context, event types.EventWithAttributeMap) error {
	sm.Logger(ctx).Info("transferred", "event", event)

	data := types.NftTransferEventData{}
	if err := json.Unmarshal([]byte(event.AttributesMap["data"]), &data); err != nil {
		return errors.New("failed to unmarshal transfer event")
	}

	objectAddr, err := getVMAddress(sm.ac, data.Object)
	if err != nil {
		return errors.New("failed to parse object address")
	}
	objectSdkAddr := getCosmosAddress(objectAddr)

	fromAddr, err := movetypes.AccAddressFromString(sm.ac, data.From)
	if err != nil {
		return errors.New("failed to parse prev owner address")
	}
	fromSdkAddr := getCosmosAddress(fromAddr)

	toAddr, err := getVMAddress(sm.ac, data.To)
	if err != nil {
		return errors.New("failed to parse new owner address")
	}
	toSdkAddr := getCosmosAddress(toAddr)

	tpk, err := sm.tokenMap.Indexes.TokenAddress.MatchExact(ctx, objectSdkAddr)
	if err != nil {
		return errors.New("token's object address not found")
	}

	token, err := sm.tokenMap.Get(ctx, tpk)
	if err != nil {
		// NOT all transferEvent means the nft is transferred. it's all object transfer event. so it's okay to ignore NotFound error
		if cosmoserr.IsOf(err, collections.ErrNotFound) {
			sm.Logger(ctx).Debug("nft not found, maybe not NFT related object transfer", "object", objectSdkAddr.String(), "prevOwner", fromSdkAddr.String())
			return nil
		}
		sm.Logger(ctx).Info("failed to get nft from prev owner and object address", "err", err, "object", objectSdkAddr.String(), "prev", fromSdkAddr.String())

		return err
	}
	token.OwnerAddr = toSdkAddr.String()

	if err = sm.tokenMap.Set(ctx, tpk, token); err != nil {
		return errors.New("failed to delete nft from sender's collection")
	}

	err = sm.applyCollectionOwnerMap(ctx, tpk.K1(), fromSdkAddr, false)
	if err != nil {
		return errors.New("failed to decrease collection count from prev owner")

	}
	err = sm.applyCollectionOwnerMap(ctx, tpk.K1(), toSdkAddr, true)
	if err != nil {
		return errors.New("failed to increase collection count from new owner")
	}

	err = sm.tokenOwnerMap.Remove(ctx, collections.Join3(fromSdkAddr, tpk.K1(), tpk.K2()))
	if err != nil {
		sm.Logger(ctx).Error("failed to remove from tokenOwnerSet", "to", toSdkAddr, "collection-addr", tpk.K1(), "token-id", tpk.K2(), "error", err)
		return errors.New("failed to insert token into tokenOwnerSet")
	}
	err = sm.tokenOwnerMap.Set(ctx, collections.Join3(toSdkAddr, tpk.K1(), tpk.K2()), true)
	if err != nil {
		sm.Logger(ctx).Error("failed to insert into tokenOwnerSet", "to", toSdkAddr, "collection-addr", tpk.K1(), "token-id", tpk.K2(), "error", err)
		return errors.New("failed to insert token into tokenOwnerSet")
	}

	sm.Logger(ctx).Info("nft transferred", "objectKey", tpk, "token", token, "prevOwner", data.From, "newOwner", data.To)
	return nil
}

func (sm MoveNftSubmodule) handleMutateEvent(ctx context.Context, event types.EventWithAttributeMap) error {
	sm.Logger(ctx).Info("mutated", "event", event)
	cdc := sm.ac

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

		nft, err := sm.getIndexedTokenFromVMStore(ctx, objectAddr, nil)
		if err != nil {
			return errors.New("failed to get minted nft info")
		}
		sm.Logger(ctx).Debug("mutated", "nft", nft)

		// remove the nft from the sender's collection
		tpk, err := sm.tokenMap.Indexes.TokenAddress.MatchExact(ctx, objectSdkAddr)
		if err != nil {
			return errors.New("object key not found")
		}
		nft.CollectionAddr = tpk.K1().String()

		if err = sm.tokenMap.Set(ctx, tpk, *nft); err != nil {
			return errors.New("failed to update mutated nft")
		}
	case data.Collection != "":
		collectionAddr, err := getVMAddress(cdc, data.Collection)
		if err != nil {
			return errors.New("failed to parse object address")
		}

		collection, err := sm.getIndexedCollectionFromVMStore(ctx, collectionAddr)
		if err != nil {
			return errors.New("failed to get mutated collection info")
		}

		err = sm.collectionMap.Set(ctx, getCosmosAddress(collectionAddr), *collection)
		if err != nil {
			return errors.New("failed to update mutated collection")
		}
	}

	sm.Logger(ctx).Info("nft mutated", "nft", data.Nft, "collection", data.Collection)

	return nil
}

func (sm MoveNftSubmodule) handleBurnEvent(ctx context.Context, event types.EventWithAttributeMap) error {
	sm.Logger(ctx).Info("burnt", "event", event)
	cdc := sm.ac
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
	tpk, err := sm.tokenMap.Indexes.TokenAddress.MatchExact(ctx, objectSdkAddr)
	if err != nil {
		return errors.New("token's object address not found")
	}
	token, err := sm.tokenMap.Get(ctx, tpk)
	if err != nil {
		return errors.New("failed to get nft from tokenMap")
	}

	err = sm.tokenMap.Remove(ctx, tpk)
	if err != nil {
		return errors.New("failed to delete nft from tokenMap")
	}

	collectionAddr, _ := getVMAddress(cdc, token.CollectionAddr)
	collectionSdkAddr := getCosmosAddress(collectionAddr)

	ownerAddr, _ := getVMAddress(cdc, token.OwnerAddr)
	ownerSdkAddr := getCosmosAddress(ownerAddr)

	err = sm.applyCollectionOwnerMap(ctx, collectionSdkAddr, ownerSdkAddr, false)
	if err != nil {
		return err // just return err, no wrap
	}

	err = sm.tokenOwnerMap.Remove(ctx, collections.Join3(ownerSdkAddr, tpk.K1(), tpk.K2()))
	if err != nil {
		sm.Logger(ctx).Error("failed to remove from tokenOwnerSet", "owner", ownerSdkAddr, "collection-addr", tpk.K1(), "token-id", tpk.K2(), "error", err)
		return errors.New("failed to insert token into tokenOwnerSet")
	}

	sm.Logger(ctx).Info("nft burnt", "event", burnt)
	return nil
}

func (sm MoveNftSubmodule) applyCollectionOwnerMap(ctx context.Context, collectionAddr, ownerAddr sdk.AccAddress, isIncrease bool) error {
	count, err := sm.collectionOwnerMap.Get(ctx, collections.Join(ownerAddr, collectionAddr))
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
		err = sm.collectionOwnerMap.Remove(ctx, collections.Join(ownerAddr, collectionAddr))
	} else {
		err = sm.collectionOwnerMap.Set(ctx, collections.Join(ownerAddr, collectionAddr), count)
	}
	if err != nil {
		return errors.New("failed to update collection count in collectionOwnersMap")
	}
	return nil
}
