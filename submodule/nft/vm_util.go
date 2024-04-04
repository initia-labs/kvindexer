package nft

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strconv"

	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/kvindexer/module/keeper"
	"github.com/initia-labs/kvindexer/submodule/nft/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const eventType = "wasm"

var (
	qreqCollectionContractInfo = []byte("{\"contract_info\":{}}") // {"contract_info":{}}
	qreqCollectionMinter       = []byte("{\"minter\":{}}")        // {"minter":{}}
	qreqCollectionNumTokens    = []byte("{\"num_tokens\":{}}")    // {"num_tokens":{}}
)

func encode(req []byte) []byte {
	res := make([]byte, base64.StdEncoding.EncodedLen(len(req)))
	base64.StdEncoding.Encode(res, req)
	return res
}

func generateQueryRequestToGetNftInfo(tokenId string) []byte {
	return []byte(`{"nft_info":{"token_id":"` + tokenId + `"}}`)
	//return encode(qb)
}

func getCollectionContractInfo(k *keeper.Keeper, ctx context.Context, colAddr sdk.AccAddress) (*types.ContractInfo, error) {
	rb, err := k.VMKeeper.QuerySmart(ctx, colAddr, []byte("{\"contract_info\":{}}")) //qreqCollectionContractInfo)
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

	q := generateQueryRequestToGetNftInfo(tokenId)
	rb, err := k.VMKeeper.QuerySmart(ctx, collectionAddr, q)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := json.Unmarshal(rb, &resource); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &resource, nil
}

func getIndexedNftFromVMStore(k *keeper.Keeper, ctx context.Context, contractAddr sdk.AccAddress, tokenId string, ownerAddr *sdk.AccAddress) (*types.IndexedToken, error) {
	resource, err := getNftResourceFromVMStore(k, ctx, contractAddr, tokenId)
	if err != nil {
		return nil, err
	}
	indexed := types.IndexedToken{
		Nft: &types.Token{
			TokenId: tokenId,
			Uri:     resource.TokenUri,
		},
		CollectionAddr: contractAddr.String(),
	}
	if ownerAddr != nil {
		indexed.OwnerAddr = ownerAddr.String()
	}

	return &indexed, nil
}

// wasm only uses bech32 address, not hex
func getVMAddress(_ address.Codec, addr string) (sdk.AccAddress, error) {
	return sdk.AccAddressFromBech32(addr)
}

// it just returns the same address: it's like an abstract function to support other VMs
func getCosmosAddress(addr sdk.AccAddress) sdk.AccAddress {
	return addr
}
