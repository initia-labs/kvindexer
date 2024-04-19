package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/initia-labs/OPinit/x/opchild/types"
)

type OPChildKeeper interface {
	GetValidatorByConsAddr(ctx context.Context, addr sdk.ConsAddress) (types.Validator, bool)
}
