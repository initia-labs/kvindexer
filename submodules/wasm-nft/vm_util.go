package wasm_nft

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strconv"

	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	nfttypes "github.com/initia-labs/kvindexer/nft/types"
	"github.com/initia-labs/kvindexer/submodules/move-nft/types"
)

var eventTypes = []string{"wasm"}

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

func (sm WasmNFTSubmodule) getCollectionContractInfo(ctx context.Context, colAddr sdk.AccAddress) (*types.ContractInfo, error) {
	rb, err := sm.vmKeeper.QuerySmart(ctx, colAddr, []byte("{\"contract_info\":{}}")) //qreqCollectionContractInfo)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := types.ContractInfo{}
	if err := json.Unmarshal(rb, &res); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &res, nil
}

func (sm WasmNFTSubmodule) getCollectionMinter(ctx context.Context, colAddr sdk.AccAddress) (*types.Minter, error) {
	rb, err := sm.vmKeeper.QuerySmart(ctx, colAddr, qreqCollectionMinter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := types.Minter{}
	if err := json.Unmarshal(rb, &res); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &res, nil
}

func (sm WasmNFTSubmodule) getCollectionNumTokens(ctx context.Context, colAddr sdk.AccAddress) (*types.NumTokens, error) {
	rb, err := sm.vmKeeper.QuerySmart(ctx, colAddr, qreqCollectionNumTokens)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := types.NumTokens{}
	if err := json.Unmarshal(rb, &res); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &res, nil
}

func (sm WasmNFTSubmodule) getCollectionFromVMStore(ctx context.Context, colAddr sdk.AccAddress) (*types.CollectionResource, error) {
	resource := types.CollectionResource{}

	ci, err := sm.getCollectionContractInfo(ctx, colAddr)
	if err != nil {
		return nil, err
	}
	resource.Collection.Name = ci.Name

	minter, err := sm.getCollectionMinter(ctx, colAddr)
	if err != nil {
		return nil, err
	}
	resource.Collection.Creator = minter.Minter

	numTokens, err := sm.getCollectionNumTokens(ctx, colAddr)
	if err != nil {
		return nil, err
	}
	resource.Collection.Nfts = &nfttypes.TokenHandle{Length: strconv.FormatInt(numTokens.Count, 10)}

	return &resource, nil
}

func (sm WasmNFTSubmodule) getIndexedCollectionFromVMStore(ctx context.Context, colAddr sdk.AccAddress) (*nfttypes.IndexedCollection, error) {
	resource, err := sm.getCollectionFromVMStore(ctx, colAddr)
	if err != nil {
		return nil, err
	}
	indexed := nfttypes.IndexedCollection{
		Collection: &resource.Collection,
		ObjectAddr: colAddr.String(),
	}
	return &indexed, nil
}

func (sm WasmNFTSubmodule) getNftResourceFromVMStore(ctx context.Context, collectionAddr sdk.AccAddress, tokenId string) (*types.NftResource, error) {
	resource := types.NftResource{}

	q := generateQueryRequestToGetNftInfo(tokenId)
	rb, err := sm.vmKeeper.QuerySmart(ctx, collectionAddr, q)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := json.Unmarshal(rb, &resource); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &resource, nil
}

func (sm WasmNFTSubmodule) getIndexedNftFromVMStore(ctx context.Context, contractAddr sdk.AccAddress, tokenId string, ownerAddr *sdk.AccAddress) (*nfttypes.IndexedToken, error) {
	resource, err := sm.getNftResourceFromVMStore(ctx, contractAddr, tokenId)
	if err != nil {
		return nil, err
	}
	indexed := nfttypes.IndexedToken{
		Nft: &nfttypes.Token{
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
