package evm_nft

import (
	"context"
	"encoding/base64"
	"strings"

	"cosmossdk.io/core/address"
	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	nfttypes "github.com/initia-labs/kvindexer/nft/types"
	"github.com/initia-labs/kvindexer/submodules/evm-nft/types"
)

var eventTypes = []string{"evm"}

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

func (sm EvmNFTSubmodule) getCollectionMinter(ctx context.Context, classId string) (*types.Minter, error) {
	return nil, errors.New("not supported")
}

func (sm EvmNFTSubmodule) getCollectionNumTokens(ctx context.Context, colAddr sdk.AccAddress) (*types.NumTokens, error) {
	return nil, errors.New("not supported")
}

func (sm EvmNFTSubmodule) getCollectionFromVMStore(ctx context.Context, classId string) (*types.CollectionResource, error) {
	resource := types.CollectionResource{}

	className, classUri, classData, err := sm.vmKeeper.ERC721Keeper().GetClassInfo(ctx, classId)
	sm.Logger(ctx).Info("[DEBUG] getCollectionContractInfo", "className", className, "classUri", classUri, "classData", classData, "error", err)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resource.Collection.Name = className
	resource.Collection.Uri = classUri
	resource.Collection.Description = classData

	/* unavailable in evm
	minter, err := sm.getCollectionMinter(ctx, classId)
	if err != nil {
		return nil, err
	}
	resource.Collection.Creator = minter.Minter
	numTokens, err := sm.getCollectionNumTokens(ctx, classId)
	if err != nil {
		return nil, err
	}
	resource.Collection.Nfts = &nfttypes.TokenHandle{Length: strconv.FormatInt(numTokens.Count, 10)}
	*/

	return &resource, nil
}

func (sm EvmNFTSubmodule) getIndexedCollectionFromVMStore(ctx context.Context, contractAddress common.Address, classId string) (*nfttypes.IndexedCollection, error) {
	resource, err := sm.getCollectionFromVMStore(ctx, classId)
	if err != nil {
		return nil, err
	}

	contractSdkAddr, err := getCosmosAddressFromString(sm.ac, contractAddress.Hex())
	if err != nil {
		sm.Logger(ctx).Error("[DEBUG] contractAddress", "str", contractAddress.String(), "bytes", contractAddress.Bytes(), "hex", contractAddress.Hex())
		return nil, errors.Wrap(err, "invalid contract address")
	}

	indexed := nfttypes.IndexedCollection{
		Collection: &resource.Collection,
		ObjectAddr: contractSdkAddr.String(),
	}
	return &indexed, nil
}

func (sm EvmNFTSubmodule) getNftResourceFromVMStore(ctx context.Context, classId, tokenId string) (*types.NftResource, error) {
	tokenUris, _, err := sm.vmKeeper.ERC721Keeper().GetTokenInfos(ctx, classId, []string{tokenId})
	if err != nil {
		sm.Logger(ctx).Warn("DEBUG] getNftResourceFromVMStore", "classId", classId, "tokenId", tokenId, "error", err)
		return nil, errors.Wrap(err, "failed to get token info")
	}
	resource := types.NftResource{}
	resource.TokenUri = tokenUris[0]

	return &resource, nil
}

func (sm EvmNFTSubmodule) getIndexedNftFromVMStore(ctx context.Context, contractAddress common.Address, classId, tokenId string, ownerAddr *sdk.AccAddress) (*nfttypes.IndexedToken, error) {
	resource, err := sm.getNftResourceFromVMStore(ctx, classId, tokenId)
	if err != nil {
		return nil, err
	}
	indexed := nfttypes.IndexedToken{
		Nft: &nfttypes.Token{
			TokenId: tokenId,
			Uri:     resource.TokenUri,
		},
		CollectionAddr: getCosmosAddress(contractAddress).String(),
	}
	if ownerAddr != nil {
		indexed.OwnerAddr = ownerAddr.String()
	}

	return &indexed, nil
}

func getCosmosAddress(addr common.Address) sdk.AccAddress {
	return sdk.AccAddress(addr.Bytes())
}

func getCosmosAddressFromString(ac address.Codec, addr string) (sdk.AccAddress, error) {
	addr = strings.ToLower(addr)
	if sdkAddr, err := sdk.AccAddressFromBech32(addr); err == nil {
		return sdkAddr, nil
	}
	return sdk.AccAddressFromHexUnsafe(strings.TrimPrefix(strings.TrimPrefix(addr, "0x"), "000000000000000000000000"))
}
