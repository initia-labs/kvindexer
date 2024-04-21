package types

import (
	"context"

	nfttransfertypes "github.com/initia-labs/initia/x/ibc/nft-transfer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type NftTransferKeeper interface {
	GetAllClassTraces(ctx context.Context) (nfttransfertypes.Traces, error)
}

type WasmKeeper interface {
	QuerySmart(ctx context.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error)
}

type PairSubmodule interface {
	GetPair(ctx context.Context, isFungible bool, l2key string) (string, error)
}
