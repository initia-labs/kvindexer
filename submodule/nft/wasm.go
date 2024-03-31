//go:build vm_wasm

package nft

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strconv"

	"cosmossdk.io/core/address"
	cosmoserr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/indexer/v2/config"
	"github.com/initia-labs/indexer/v2/module/keeper"
	"github.com/initia-labs/indexer/v2/submodule/nft/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const eventType = "wasm"

func processEvents(k *keeper.Keeper, ctx context.Context, cfg config.SubmoduleConfig, events []types.EventWithAttributeMap) error {
	var fn func(k *keeper.Keeper, ctx context.Context, cfg config.SubmoduleConfig, event types.EventWithAttributeMap) error
	for _, event := range events {
		switch event.AttributesMap["action"] {
		case "mint":
			fn = handleMintEvent
		case "transfer_nft":
			fn = handlerTransferEvent
		case "send_nft":
			fn = handleSendEvent
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

	//data := types.MintEvent{}

	// k.Logger(ctx).Info("nft minted", "collection", collection, "nft", token, "collection-sdk-addr", collectionSdkAddr, "nft-sdk-addr", tokenSdkAddr, "creator-sdk-addr", creatorSdkAddr)
	return nil
}

func handlerTransferEvent(k *keeper.Keeper, ctx context.Context, cfg config.SubmoduleConfig, event types.EventWithAttributeMap) error {
	k.Logger(ctx).Info("transferred", "event", event)

	panic("not implemented")

	//k.Logger(ctx).Debug("nft transferred", "objectKey", "token", token, "prevOwner", data.From, "newOwner", data.To)
	return nil
}

func handleSendEvent(k *keeper.Keeper, ctx context.Context, cfg config.SubmoduleConfig, event types.EventWithAttributeMap) error {
	k.Logger(ctx).Info("mutated", "event", event)
	//cdc := k.GetAddressCodec()

	panic("not implemented")

	//k.Logger(ctx).Debug("nft mutated", "nft", data.Nft, "collection", data.Collection)

	return nil
}
func handleBurnEvent(k *keeper.Keeper, ctx context.Context, cfg config.SubmoduleConfig, event types.EventWithAttributeMap) error {
	k.Logger(ctx).Info("burnt", "event", event)

	panic("not implemented")

	//k.Logger(ctx).Debug("nft burnt", "objectKey", objectKey, "nft", burnt.Nft)

	//return nftByOwner.Remove(ctx, objectKey)
	return nil
}

var (
	qreqCollectionContractInfo = []byte("eyJjb250cmFjdF9pbmZvIjp7fX0=") // {"contract_info":{}}
	qreqCollectionMinter       = []byte("eyJtaW50ZXIiOnt9fQ==")         // {"minter":{}}
	qreqCollectionNumTokens    = []byte("eyJudW1fdG9rZW5zIjp7fX0=")     // {"num_tokens":{}}
)

func encode(req []byte) []byte {
	res := make([]byte, base64.StdEncoding.EncodedLen(len(req)))
	base64.StdEncoding.Encode(res, req)
	return res
}

func generateQueryRequestToGetNftInfo(tokenId string) []byte {
	qb := []byte(`{"nft_info":{"token_id":"` + tokenId + `"}}`)
	return encode(qb)
}

func getCollectionContractInfo(k *keeper.Keeper, ctx context.Context, colAddr sdk.AccAddress) (*types.ContractInfo, error) {
	rb, err := k.VMKeeper.QuerySmart(ctx, colAddr, qreqCollectionContractInfo)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := types.ContractInfo{}
	if err := json.Unmarshal(rb, &res); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &res, nil
}

func getCollectioMinter(k *keeper.Keeper, ctx context.Context, colAddr sdk.AccAddress) (*types.Minter, error) {
	rb, err := k.VMKeeper.QuerySmart(ctx, colAddr, qreqCollectionMinter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := types.Minter{}
	if err := json.Unmarshal(rb, &res); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &res, nil
}

func getCollectioNumTokens(k *keeper.Keeper, ctx context.Context, colAddr sdk.AccAddress) (*types.NumTokens, error) {
	rb, err := k.VMKeeper.QuerySmart(ctx, colAddr, qreqCollectionNumTokens)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := types.NumTokens{}
	if err := json.Unmarshal(rb, &res); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &res, nil
}

func getCollectionFromVMStore(k *keeper.Keeper, ctx context.Context, colAddr sdk.AccAddress) (*types.CollectionResource, error) {
	resource := types.CollectionResource{}

	ci, err := getCollectionContractInfo(k, ctx, colAddr)
	if err != nil {
		return nil, err
	}
	resource.Collection.Name = ci.Name

	minter, err := getCollectioMinter(k, ctx, colAddr)
	if err != nil {
		return nil, err
	}
	resource.Collection.Creator = minter.Minter

	numTokens, err := getCollectioNumTokens(k, ctx, colAddr)
	if err != nil {
		return nil, err
	}
	resource.Collection.Nfts = &types.TokenHandle{Length: strconv.FormatInt(numTokens.Count, 10)}

	return &resource, nil
}

func getIndexedCollectionFromVMStore(k *keeper.Keeper, ctx context.Context, colAddr sdk.AccAddress) (*types.IndexedCollection, error) {
	resource, err := getCollectionFromVMStore(k, ctx, colAddr)
	if err != nil {
		return nil, err
	}
	indexed := types.IndexedCollection{
		Collection: &resource.Collection,
		ObjectAddr: colAddr.String(),
	}
	return &indexed, nil
}

func getNftResourceFromVMStore(k *keeper.Keeper, ctx context.Context, collectionAddr sdk.AccAddress, tokenId string) (*types.NftResource, error) {
	resource := types.NftResource{}

	panic("not implemented")

	return &resource, nil
}

func getIndexedNftFromVMStore(k *keeper.Keeper, ctx context.Context, nftAddr sdk.AccAddress, collectionAddr *sdk.AccAddress) (*types.IndexedToken, error) {
	/* FIXME: implement this
	resource, err := getNftResourceFromVMStore(k, ctx, nftAddr)
	if err != nil {
		return nil, err
	}
	indexed := types.IndexedNft{
		ObjectAddr: nftAddr.String(),
		Nft:        &resource.Nft,
	}
	if collectionAddr != nil {
		indexed.CollectionAddr = collectionAddr.String()
	}

	return &indexed, nil
	*/
	return nil, nil
}

func getVMAddress(cdc address.Codec, addr string) (sdk.AccAddress, error) {
	return sdk.AccAddressFromBech32(addr)
}

// it just returns the same address: it's like an abstract function to support other VMs
func getCosmosAddress(addr sdk.AccAddress) sdk.AccAddress {
	return addr
}
