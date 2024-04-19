package types

import (
	"context"

	nfttransfertypes "github.com/initia-labs/initia/x/ibc/nft-transfer/types"

	movetypes "github.com/initia-labs/initia/x/move/types"
	vmtypes "github.com/initia-labs/movevm/types"
)

type NftTransferKeeper interface {
	GetAllClassTraces(ctx context.Context) (nfttransfertypes.Traces, error)
}

type MoveKeeper interface {
	GetResource(ctx context.Context, addr vmtypes.AccountAddress, structTag vmtypes.StructTag) (movetypes.Resource, error)
}

type PairSubmodule interface {
	GetPair(ctx context.Context, isFungible bool, l2key string) (string, error)
}
