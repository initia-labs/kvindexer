package move_nft

import (
	"context"
	"encoding/json"

	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"

	movetypes "github.com/initia-labs/initia/x/move/types"
	vmtypes "github.com/initia-labs/movevm/types"

	nfttypes "github.com/initia-labs/kvindexer/nft/types"
	"github.com/initia-labs/kvindexer/submodules/move-nft/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (sm MoveNftSubmodule) getCollectionFromVMStore(ctx context.Context, colAddr vmtypes.AccountAddress) (*types.CollectionResource, error) {
	rb, err := sm.vmKeeper.GetResource(ctx, colAddr, collectionStructTag)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	resource := types.CollectionResource{}
	if err := json.Unmarshal([]byte(rb.MoveResource), &resource); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &resource, nil
}

func (sm MoveNftSubmodule) getIndexedCollectionFromVMStore(ctx context.Context, colAddr vmtypes.AccountAddress) (*nfttypes.IndexedCollection, error) {
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

func (sm MoveNftSubmodule) getNftResourceFromVMStore(ctx context.Context, nftAddr vmtypes.AccountAddress) (*types.NftResource, error) {
	rb, err := sm.vmKeeper.GetResource(ctx, nftAddr, nftStructTag)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	resource := types.NftResource{}
	if err := json.Unmarshal([]byte(rb.MoveResource), &resource); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &resource, nil
}

func (sm MoveNftSubmodule) getIndexedTokenFromVMStore(ctx context.Context, nftAddr vmtypes.AccountAddress, collectionAddr *vmtypes.AccountAddress) (*nfttypes.IndexedToken, error) {
	resource, err := sm.getNftResourceFromVMStore(ctx, nftAddr)
	if err != nil {
		return nil, err
	}
	indexed := nfttypes.IndexedToken{
		ObjectAddr: nftAddr.String(),
		Nft:        &resource.Nft,
	}
	if collectionAddr != nil {
		indexed.CollectionAddr = collectionAddr.String()
	}

	return &indexed, nil
}

func getVMAddress(ac address.Codec, addr string) (vmtypes.AccountAddress, error) {
	accAddr, err := movetypes.AccAddressFromString(ac, addr)
	if err != nil {
		return vmtypes.AccountAddress{}, err
	}

	return vmtypes.AccountAddress(accAddr), nil
}

func getCosmosAddress(addr vmtypes.AccountAddress) sdk.AccAddress {
	return movetypes.ConvertVMAddressToSDKAddress(addr)
}
