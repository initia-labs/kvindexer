package types

import (
	"context"

	"cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/common"
	nfttransfertypes "github.com/initia-labs/initia/x/ibc/nft-transfer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type NftTransferKeeper interface {
	GetAllClassTraces(ctx context.Context) (nfttransfertypes.Traces, error)
}

type EvmNftKeeper interface {
	GetClassInfo(ctx context.Context, classId string) (className string, classUri string, classDescs string, err error)
	GetTokenInfos(ctx context.Context, classId string, tokenIds []string) (tokenUris []string, tokenDescs []string, err error)
	OwnerOf(ctx context.Context, tokenId string, classId string) (common.Address, error)
	BalanceOf(ctx context.Context, addr sdk.AccAddress, classId string) (math.Int, error)
	QuerySmart(ctx context.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error)
}

type PairSubmodule interface {
	GetPair(ctx context.Context, isFungible bool, l2key string) (string, error)
}
