package wasm_nft

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	cosmoserr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/initia-labs/kvindexer/submodules/wasm-nft/types"
)

func (sm WasmNFTSubmodule) finalizeBlock(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	sm.Logger(ctx).Debug("finalizeBlock", "submodule", types.SubmoduleName, "txs", len(req.Txs), "height", req.Height)

	for _, txResult := range res.TxResults {
		events := filterAndParseEvent(txResult.Events, eventTypes)
		err := sm.processEvents(ctx, events)
		if err != nil {
			sm.Logger(ctx).Warn("processEvents", "error", err)
		}
	}

	return nil
}

func (sm WasmNFTSubmodule) processEvents(ctx context.Context, events []types.EventWithAttributeMap) error {
	var fn func(ctx context.Context, event types.EventWithAttributeMap) error
	for _, event := range events {
		if event.Type == "wasm" {
			switch event.AttributesMap["action"] {
			case "mint":
				fn = sm.handleMintEvent
			case "transfer_nft", "send_nft":
				fn = sm.handlerSendOrTransferEvent
			case "burn":
				fn = sm.handleBurnEvent
			default:
				continue
			}
		}

		if err := fn(ctx, event); err != nil {
			sm.Logger(ctx).Error("failed to handle nft-related event", "error", err.Error())
			return cosmoserr.Wrap(err, "failed to handle nft-related event")
		}
	}
	return nil
}

func (sm WasmNFTSubmodule) handleMintEvent(ctx context.Context, event types.EventWithAttributeMap) error {
	sm.Logger(ctx).Debug("minted", "event", event)

	data := types.MintEvent{}
	if err := data.Parse(event); err != nil {
		// may be not nft mint event
		return nil
	}

	collection, err := sm.collectionMap.Get(ctx, data.ContractAddress)
	if err != nil {
		if !cosmoserr.IsOf(err, collections.ErrNotFound) {
			return cosmoserr.Wrap(err, "failed to check collection existence")
		}
		// if not found, it means this is the first minting of the collection, so we need to set into collectionMap
		coll, err := sm.getIndexedCollectionFromVMStore(ctx, data.ContractAddress)
		if err != nil {
			return cosmoserr.Wrap(err, "failed to get collection contract info")
		}
		collection = *coll

		err = sm.collectionMap.Set(ctx, data.ContractAddress, collection)
		if err != nil {
			return cosmoserr.Wrap(err, "failed to set collection")
		}
	}

	err = sm.applyCollectionOwnerMap(ctx, data.ContractAddress, data.Minter, true)
	if err != nil {
		return cosmoserr.Wrap(err, "failed to insert collection into collectionOwnersMap")
	}

	token, err := sm.getIndexedNftFromVMStore(ctx, data.ContractAddress, data.TokenId, &data.Minter)
	if err != nil {
		return cosmoserr.Wrap(err, "failed to get token info")
	}
	token.CollectionName = collection.Collection.Name

	err = sm.tokenMap.Set(ctx, collections.Join(data.ContractAddress, data.TokenId), *token)
	if err != nil {
		return cosmoserr.Wrap(err, "failed to set token")
	}

	err = sm.tokenOwnerMap.Set(ctx, collections.Join3(data.Minter, data.ContractAddress, data.TokenId), true)
	if err != nil {
		sm.Logger(ctx).Error("failed to insert into tokenOwnerSet", "minter", data.Minter, "collection-addr", data.ContractAddress, "token-id", token.Nft.TokenId, "error", err)
		return cosmoserr.Wrap(err, "failed to insert into tokenOwnerSet")
	}

	sm.Logger(ctx).Info("nft minted", "collection", collection, "token", token)
	return nil
}

func (sm WasmNFTSubmodule) handlerSendOrTransferEvent(ctx context.Context, event types.EventWithAttributeMap) (err error) {
	sm.Logger(ctx).Info("sent/transferred", "event", event)
	data := types.TransferOrSendEvent{}
	if err := data.Parse(event); err != nil {
		// may be not nft send/transfer event
		return nil
	}

	tpk := collections.Join[sdk.AccAddress, string](data.ContractAddress, data.TokenId)

	token, err := sm.tokenMap.Get(ctx, tpk)
	if err != nil {
		sm.Logger(ctx).Debug("failed to get nft from prev owner and object addres", "collection-addr", data.ContractAddress, "token-id", data.TokenId, "prevOwner", data.Sender, "error", err)
		return cosmoserr.Wrap(err, "failed to get nft from tokenMap")
	}
	token.OwnerAddr = data.Recipient.String()

	if err = sm.tokenMap.Set(ctx, tpk, token); err != nil {
		return errors.New("failed to delete nft from sender's collection")
	}

	err = sm.applyCollectionOwnerMap(ctx, tpk.K1(), data.Sender, false)
	if err != nil {
		return errors.New("failed to decrease collection count from prev owner")

	}
	err = sm.applyCollectionOwnerMap(ctx, tpk.K1(), data.Recipient, true)
	if err != nil {
		return errors.New("failed to increase collection count from new owner")
	}

	err = sm.tokenOwnerMap.Remove(ctx, collections.Join3(data.Sender, tpk.K1(), tpk.K2()))
	if err != nil {
		sm.Logger(ctx).Error("failed to remove from tokenOwnerSet", "to", data.Recipient, "collection-addr", tpk.K1(), "token-id", tpk.K2(), "error", err)
		return errors.New("failed to insert token into tokenOwnerSet")
	}
	err = sm.tokenOwnerMap.Set(ctx, collections.Join3(data.Recipient, tpk.K1(), tpk.K2()), true)
	if err != nil {
		sm.Logger(ctx).Error("failed to insert into tokenOwnerSet", "to", data.Recipient, "collection-addr", tpk.K1(), "token-id", tpk.K2(), "error", err)
		return errors.New("failed to insert token into tokenOwnerSet")
	}

	sm.Logger(ctx).Info("nft sent/transferred", "objectKey", tpk, "token", token, "prevOwner", data.Sender, "newOwner", data.Recipient)
	return nil
}

func (sm WasmNFTSubmodule) handleBurnEvent(ctx context.Context, event types.EventWithAttributeMap) error {
	sm.Logger(ctx).Info("burnt", "event", event)

	data := types.BurnEvent{}
	if err := data.Parse(event); err != nil {
		// may be not nft burn event
		return nil
	}

	// remove from tokensOwnersMap
	tpk := collections.Join[sdk.AccAddress, string](data.ContractAddress, data.TokenId)
	token, err := sm.tokenMap.Get(ctx, tpk)
	if err != nil {
		return cosmoserr.Wrap(err, "failed to get nft from tokenMap")
	}

	err = sm.tokenMap.Remove(ctx, tpk)
	if err != nil {
		return cosmoserr.Wrap(err, "failed to delete nft from tokenMap")
	}

	ownerAddr, _ := getVMAddress(sm.ac, token.OwnerAddr)
	ownerSdkAddr := getCosmosAddress(ownerAddr)

	err = sm.tokenOwnerMap.Set(ctx, collections.Join3(ownerSdkAddr, tpk.K1(), tpk.K2()), true)
	if err != nil {
		sm.Logger(ctx).Error("failed to remove from tokenOwnerSet", "owner", ownerSdkAddr, "collection-addr", tpk.K1(), "token-id", tpk.K2(), "error", err)
		return cosmoserr.Wrap(err, "failed to insert token into tokenOwnerSet")
	}

	err = sm.applyCollectionOwnerMap(ctx, data.ContractAddress, ownerSdkAddr, false)
	if err != nil {
		return err // just return err, no wrap
	}

	sm.Logger(ctx).Info("nft burnt", "event", data)

	return nil
}

func (sm WasmNFTSubmodule) applyCollectionOwnerMap(ctx context.Context, collectionAddr, ownerAddr sdk.AccAddress, isIncrease bool) error {
	count, err := sm.collectionOwnerMap.Get(ctx, collections.Join(ownerAddr, collectionAddr))
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
		err = sm.collectionOwnerMap.Remove(ctx, collections.Join(ownerAddr, collectionAddr))
	} else {
		err = sm.collectionOwnerMap.Set(ctx, collections.Join(ownerAddr, collectionAddr), count)
	}
	if err != nil {
		return cosmoserr.Wrap(err, "failed to update collection count in collectionOwnersMap")
	}
	return nil
}
