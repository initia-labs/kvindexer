package types

import (
	"encoding/json"
	"errors"

	"cosmossdk.io/core/address"

	sdk "github.com/cosmos/cosmos-sdk/types"

	evmtypes "github.com/initia-labs/minievm/x/evm/types"
)

type TransferLog evmtypes.Log
type ParsedTransfer struct {
	Address sdk.AccAddress
	From    sdk.AccAddress
	To      sdk.AccAddress
	TokenId string
}

func (tl TransferLog) IsErc721Transfer() bool {
	if len(tl.Topics) != 4 {
		return false
	}
	return tl.Topics[0] == "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
}

func ParseERC721TransferLog(ac address.Codec, attributeValue string) (parsed *ParsedTransfer, err error) {
	tl := TransferLog{}
	err = json.Unmarshal([]byte(attributeValue), &tl)
	if err != nil {
		return nil, errors.New("the attribute is not log")
	}
	if !tl.IsErc721Transfer() {
		return nil, errors.New("not erc721 transfer")
	}

	addr, err := evmtypes.ContractAddressFromString(ac, tl.Address)
	if err != nil {
		return nil, errors.New("invalid contract address")
	}
	from, err := evmtypes.ContractAddressFromString(ac, tl.Topics[1])
	if err != nil {
		return nil, errors.New("invalid from address")
	}
	to, err := evmtypes.ContractAddressFromString(ac, tl.Topics[2])
	if err != nil {
		return nil, errors.New("invalid to address")
	}

	return &ParsedTransfer{
		Address: sdk.AccAddress(addr[:]),
		From:    sdk.AccAddress(from[:]),
		To:      sdk.AccAddress(to[:]),
		TokenId: tl.Topics[3],
	}, nil
}

type NftAction int

const (
	NftActionMint = NftAction(0x1000 + iota)
	NftActionBurn
	NftActionTransfer
)

func (pt ParsedTransfer) GetAction() NftAction {
	if pt.From.Empty() && !pt.To.Empty() {
		return NftActionMint
	}
	if !pt.From.Empty() && pt.To.Empty() {
		return NftActionBurn
	}
	return NftActionTransfer
}
