package types

import (
	"context"
	"github.com/initia-labs/OPinit/x/opchild/types"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	distrtypes "github.com/initia-labs/initia/x/distribution/types"
	nfttransfertypes "github.com/initia-labs/initia/x/ibc/nft-transfer/types"
	stakingtypes "github.com/initia-labs/initia/x/mstaking/types"
	rewardtypes "github.com/initia-labs/initia/x/reward/types"

	oracletypes "github.com/skip-mev/slinky/x/oracle/types"
)

// AccountKeeper is expected keeper for auth module
//type AccountKeeper interface {
//	NewAccount(ctx context.Context, acc sdk.AccountI) sdk.AccountI
//	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
//	SetAccount(ctx context.Context, acc sdk.AccountI)
//	HasAccount(ctx context.Context, addr sdk.AccAddress) bool
//
//	GetModuleAddress(name string) sdk.AccAddress
//	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
//	SetModuleAccount(ctx context.Context, macc sdk.ModuleAccountI)
//
//	NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
//	NextAccountNumber(ctx context.Context) uint64
//
//	IterateAccounts(ctx context.Context, cb func(account sdk.AccountI) (stop bool))
//}

// BankViewKeeper defines a subset of methods implemented by the cosmos-sdk bank keeper
type BankViewKeeper interface {
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// BankKeeper defines a subset of methods implemented by the cosmos-sdk bank keeper
type BankKeeper interface {
	BankViewKeeper
	IsSendEnabledCoins(ctx context.Context, coins ...sdk.Coin) error
	BlockedAddr(addr sdk.AccAddress) bool
	GetSupply(ctx context.Context, denom string) sdk.Coin
	HasSupply(ctx context.Context, denom string) bool
	GetPaginatedTotalSupply(ctx context.Context, pagination *query.PageRequest) (sdk.Coins, *query.PageResponse, error)
	IterateTotalSupply(ctx context.Context, cb func(sdk.Coin) bool)
	//GetDenomMetaData(ctx context.Context, denom string) (types.Metadata, bool)
	HasDenomMetaData(ctx context.Context, denom string) bool
	//GetAllDenomMetaData(ctx context.Context) []types.Metadata
	//IterateAllDenomMetaData(ctx context.Context, cb func(types.Metadata) bool)
}

type RewardKeeper interface {
	GetParams(ctx context.Context) (params rewardtypes.Params, err error)
}

// StakingKeeper is expected keeper for staking module
type StakingKeeper interface {
	Validator(ctx context.Context, address sdk.ValAddress) (stakingtypes.ValidatorI, error)
	GetBondedValidatorsByPower(ctx context.Context) ([]stakingtypes.Validator, error)
	UnbondingTime(ctx context.Context) (time.Duration, error)
	Delegate(ctx context.Context, delAddr sdk.AccAddress, bondAmt sdk.Coins, tokenSrc stakingtypes.BondStatus, validator stakingtypes.Validator, subtractAccount bool) (sdk.DecCoins, error)
	GetValidator(ctx context.Context, addr sdk.ValAddress) (stakingtypes.Validator, error)
	Unbond(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, shares sdk.DecCoins) (amount sdk.Coins, err error)
	Delegation(context.Context, sdk.AccAddress, sdk.ValAddress) (stakingtypes.DelegationI, error)
	BondDenoms(ctx context.Context) (res []string, err error)
	SetBondDenoms(ctx context.Context, bondDenoms []string) error
}

// DistributionKeeper is expected keeper for distribution module
type DistributionKeeper interface {
	WithdrawDelegationRewards(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (distrtypes.Pools, error)

	// increment validator period, returning the period just ended
	IncrementValidatorPeriod(ctx context.Context, val stakingtypes.ValidatorI) (uint64, error)
	// calculate the total rewards accrued by a delegation
	CalculateDelegationRewards(ctx context.Context, val stakingtypes.ValidatorI, del stakingtypes.DelegationI, endingPeriod uint64) (rewards distrtypes.DecPools, err error)

	GetRewardWeights(ctx context.Context) (rewardWeights []distrtypes.RewardWeight, err error)
	SetRewardWeights(ctx context.Context, rewardWeights []distrtypes.RewardWeight) error
}

type CommunityPoolKeeper interface {
	// FundCommunityPool allows an account to directly fund the community fund pool.
	FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error
}

type OracleKeeper interface {
	GetPriceForCurrencyPair(ctx sdk.Context, cp oracletypes.CurrencyPair) (oracletypes.QuotePrice, error)
}

type TransferKeeper interface {
	GetAllDenomTraces(ctx sdk.Context) transfertypes.Traces
}

type NftTransferKeeper interface {
	GetAllClassTraces(ctx context.Context) (nfttransfertypes.Traces, error)
}

type OPChildKeeper interface {
	GetValidatorByConsAddr(ctx context.Context, addr sdk.ConsAddress) (types.Validator, bool)
}
