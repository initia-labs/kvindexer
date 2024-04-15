package nft

import (
	"context"

	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	movetypes "github.com/initia-labs/initia/x/move/types"
	"github.com/initia-labs/kvindexer/module/keeper"
	"github.com/initia-labs/kvindexer/submodule/nft/types"
	vmtypes "github.com/initia-labs/movevm/types"
)

const eventType = "move"

var collectionStructTag = vmtypes.StructTag{
	Address: vmtypes.StdAddress,
	Module:  "collection",
	Name:    "Collection",
}
var nftStructTag = vmtypes.StructTag{
	Address: vmtypes.StdAddress,
	Module:  "nft",
	Name:    "Nft",
}

func getCollectionFromVMStore(k *keeper.Keeper, ctx context.Context, colAddr vmtypes.AccountAddress) (*types.CollectionResource, error) {

	panic("not implemented")

}

func getIndexedCollectionFromVMStore(k *keeper.Keeper, ctx context.Context, colAddr vmtypes.AccountAddress) (*types.IndexedCollection, error) {
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

func getNftResourceFromVMStore(k *keeper.Keeper, ctx context.Context, nftAddr vmtypes.AccountAddress) (*types.NftResource, error) {

	panic("not implemented")

}

func getIndexedTokenFromVMStore(k *keeper.Keeper, ctx context.Context, nftAddr vmtypes.AccountAddress, collectionAddr *vmtypes.AccountAddress) (*types.IndexedToken, error) {
	resource, err := getNftResourceFromVMStore(k, ctx, nftAddr)
	if err != nil {
		return nil, err
	}
	indexed := types.IndexedToken{
		ObjectAddr: nftAddr.String(),
		Nft:        &resource.Nft,
	}
	if collectionAddr != nil {
		indexed.CollectionAddr = collectionAddr.String()
	}

	return &indexed, nil
}

func getVMAddress(cdc address.Codec, addr string) (vmtypes.AccountAddress, error) {
	accAddr, err := movetypes.AccAddressFromString(cdc, addr)
	if err != nil {
		return vmtypes.AccountAddress{}, err
	}
	return vmtypes.AccountAddress(accAddr), nil
}

func getCosmosAddress(addr vmtypes.AccountAddress) sdk.AccAddress {
	return movetypes.ConvertVMAddressToSDKAddress(addr)
}
