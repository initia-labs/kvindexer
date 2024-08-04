package evm_nft

import (
	"context"

	"github.com/pkg/errors"

	"cosmossdk.io/collections"
	cosmoserr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/initia-labs/kvindexer/submodules/evm-nft/types"
	evmtypes "github.com/initia-labs/minievm/x/evm/types"
)

func (sm EvmNFTSubmodule) finalizeBlock(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	sm.Logger(ctx).Debug("finalizeBlock", "submodule", types.SubmoduleName, "txs", len(req.Txs), "height", req.Height)

	for _, txResult := range res.TxResults {
		events := filterAndParseEvent(txResult.Events, eventTypes)
		sm.Logger(ctx).Error("[DEBUG] events", "events", events)
		err := sm.processEvents(ctx, events)
		if err != nil {
			sm.Logger(ctx).Warn("processEvents", "error", err)
		}
	}

	return nil
}

func (sm EvmNFTSubmodule) processEvents(ctx context.Context, events []types.EventWithAttributeMap) error {

	for _, event := range events {
		log, ok := event.AttributesMap[evmtypes.AttributeKeyLog]
		if !ok {
			continue // no log means it's not evm-related event
		}

		sm.Logger(ctx).Error("[DEBUG] LOGEVENT", "log", log)

		transferLog, err := types.ParseERC721TransferLog(sm.ac, log)
		if err != nil {
			sm.Logger(ctx).Debug("failed parse attribute", "error", err)
			sm.Logger(ctx).Error("[DEBUG] failed parse attribute", "error", err, "log", log)
			continue
		}
		sm.Logger(ctx).Error("[DEBUG] transferLog", "transferLog", transferLog)

		sm.Logger(ctx).Error("[DEBUG] TL", "from", transferLog.From.Bytes(), "to", transferLog.To.Bytes())

		var fn func(context.Context, *types.ParsedTransfer) error
		switch transferLog.GetAction() {
		case types.NftActionMint:
			fn = sm.handleMintEvent
		case types.NftActionTransfer:
			fn = sm.handlerTransferEvent
		case types.NftActionBurn:
			fn = sm.handleBurnEvent
		default:
			sm.Logger(ctx).Debug("unknown nft action", "action", transferLog.GetAction())
			continue
		}

		if err := fn(ctx, transferLog); err != nil {
			sm.Logger(ctx).Error("failed to handle nft-related event", "error", err.Error())
		}
	}
	return nil
}

func (sm EvmNFTSubmodule) handleMintEvent(ctx context.Context, event *types.ParsedTransfer) error {
	sm.Logger(ctx).Debug("minted", "event", event)
	sm.Logger(ctx).Error("[DEBUG] minted", "event", event)

	classId, err := evmtypes.ClassIdFromCollectionAddress(ctx, sm.vmKeeper, event.Address)
	if err != nil {
		return cosmoserr.Wrap(err, "failed to get classId from collection address")
	}
	contractSdkAddr := getCosmosAddress(event.Address)

	collection, err := sm.collectionMap.Get(ctx, contractSdkAddr)
	if err != nil {
		if !cosmoserr.IsOf(err, collections.ErrNotFound) {
			return cosmoserr.Wrap(err, "failed to check collection existence")
		}
		sm.Logger(ctx).Error("[DEBUG] collection found", "collection", collection)
		// if not found, it means this is the first minting of the collection, so we need to set into collectionMap
		coll, err := sm.getIndexedCollectionFromVMStore(ctx, event.Address, classId)
		if err != nil {
			return cosmoserr.Wrap(err, "failed to get collection contract info")
		}
		collection = *coll

		err = sm.collectionMap.Set(ctx, contractSdkAddr, collection)
		if err != nil {
			return cosmoserr.Wrap(err, "failed to set collection")
		}
	}

	err = sm.applyCollectionOwnerMap(ctx, contractSdkAddr, event.To, true)
	if err != nil {
		return cosmoserr.Wrap(err, "failed to insert collection into collectionOwnersMap")
	}

	token, err := sm.getIndexedNftFromVMStore(ctx, event.Address, event.TokenId, &event.To)
	if err != nil {
		return cosmoserr.Wrap(err, "failed to get token info")
	}
	token.CollectionName = collection.Collection.Name

	err = sm.tokenMap.Set(ctx, collections.Join(contractSdkAddr, event.TokenId), *token)
	if err != nil {
		return cosmoserr.Wrap(err, "failed to set token")
	}

	err = sm.tokenOwnerMap.Set(ctx, collections.Join3(event.To, contractSdkAddr, event.TokenId), true)
	if err != nil {
		sm.Logger(ctx).Error("failed to insert into tokenOwnerSet", "event", event, "error", err)
		return cosmoserr.Wrap(err, "failed to insert into tokenOwnerSet")
	}

	sm.Logger(ctx).Warn("nft minted", "collection", collection, "token", token)
	return nil
}

func (sm EvmNFTSubmodule) handlerTransferEvent(ctx context.Context, event *types.ParsedTransfer) (err error) {
	sm.Logger(ctx).Info("sent/transferred", "event", event)
	contractSdkAddr := getCosmosAddress(event.Address)

	tpk := collections.Join[sdk.AccAddress, string](contractSdkAddr, event.TokenId)

	token, err := sm.tokenMap.Get(ctx, tpk)
	if err != nil {
		sm.Logger(ctx).Debug("failed to get nft from prev owner and object addres", "collection-addr", event.Address, "token-id", event.TokenId, "prevOwner", event.From, "error", err)
		return cosmoserr.Wrap(err, "failed to get nft from tokenMap")
	}
	token.OwnerAddr = event.To.String()

	if err = sm.tokenMap.Set(ctx, tpk, token); err != nil {
		return errors.New("failed to delete nft from sender's collection")
	}

	err = sm.applyCollectionOwnerMap(ctx, tpk.K1(), event.From, false)
	if err != nil {
		return errors.New("failed to decrease collection count from prev owner")

	}
	err = sm.applyCollectionOwnerMap(ctx, tpk.K1(), event.To, true)
	if err != nil {
		return errors.New("failed to increase collection count from new owner")
	}

	err = sm.tokenOwnerMap.Remove(ctx, collections.Join3(event.From, tpk.K1(), tpk.K2()))
	if err != nil {
		sm.Logger(ctx).Error("failed to remove from tokenOwnerSet", "to", event.To, "collection-addr", tpk.K1(), "token-id", tpk.K2(), "error", err)
		return errors.New("failed to insert token into tokenOwnerSet")
	}
	err = sm.tokenOwnerMap.Set(ctx, collections.Join3(event.To, tpk.K1(), tpk.K2()), true)
	if err != nil {
		sm.Logger(ctx).Error("failed to insert into tokenOwnerSet", "to", event.To, "collection-addr", tpk.K1(), "token-id", tpk.K2(), "error", err)
		return errors.New("failed to insert token into tokenOwnerSet")
	}

	sm.Logger(ctx).Info("nft sent/transferred", "objectKey", tpk, "token", token, "prevOwner", event.From, "newOwner", event.To)
	return nil
}

func (sm EvmNFTSubmodule) handleBurnEvent(ctx context.Context, event *types.ParsedTransfer) error {
	sm.Logger(ctx).Info("burnt", "event", event)
	contractSdkAddr := getCosmosAddress(event.Address)

	// remove from tokensOwnersMap
	tpk := collections.Join[sdk.AccAddress, string](contractSdkAddr, event.TokenId)
	token, err := sm.tokenMap.Get(ctx, tpk)
	if err != nil {
		return cosmoserr.Wrap(err, "failed to get nft from tokenMap")
	}

	err = sm.tokenMap.Remove(ctx, tpk)
	if err != nil {
		return cosmoserr.Wrap(err, "failed to delete nft from tokenMap")
	}

	ownerSdkAddr, err := getCosmosAddressFromString(sm.ac, token.OwnerAddr)
	if err != nil {
		return cosmoserr.Wrap(err, "failed to get owner address from token")
	}

	err = sm.tokenOwnerMap.Set(ctx, collections.Join3(ownerSdkAddr, tpk.K1(), tpk.K2()), true)
	if err != nil {
		sm.Logger(ctx).Error("failed to remove from tokenOwnerSet", "owner", ownerSdkAddr, "collection-addr", tpk.K1(), "token-id", tpk.K2(), "error", err)
		return cosmoserr.Wrap(err, "failed to insert token into tokenOwnerSet")
	}

	err = sm.applyCollectionOwnerMap(ctx, contractSdkAddr, ownerSdkAddr, false)
	if err != nil {
		return err // just return err, no wrap
	}

	sm.Logger(ctx).Info("nft burnt", "event", event)

	return nil
}

func (sm EvmNFTSubmodule) applyCollectionOwnerMap(ctx context.Context, collectionAddr, ownerAddr sdk.AccAddress, isIncrease bool) error {
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
