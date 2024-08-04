package types

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"

	"cosmossdk.io/core/address"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/initia-labs/minievm/x/evm/types"
)

type NftAction int

const (
	NftActionMint = NftAction(0x1000 + iota)
	NftActionBurn
	NftActionTransfer
)

type TransferLog evmtypes.Log
type ParsedTransfer struct {
	Address common.Address
	From    sdk.AccAddress
	To      sdk.AccAddress
	TokenId string
}

//var emptyAddr sdk.AccAddress = sdk.AccAddress([]byte("0000000000000000000000000000000000000000000000000000000000000000"))

func (tl TransferLog) IsErc721Transfer() bool {
	return len(tl.Topics) == 4 && tl.Topics[0] == "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
}

func ParseERC721TransferLog(ac address.Codec, attributeValue string) (parsed *ParsedTransfer, err error) {
	tl := TransferLog{}
	err = json.Unmarshal([]byte(attributeValue), &tl)
	if err != nil {
		return nil, errors.New("the attribute is not about log")
	}
	if !tl.IsErc721Transfer() {
		return nil, errors.New("not erc721 transfer")
	}

	addr, err := evmtypes.ContractAddressFromString(ac, tl.Address)
	if err != nil {
		return nil, errors.Wrap(err, "invalid contract address")
	}

	from, err := sdk.AccAddressFromHexUnsafe(strings.TrimPrefix(tl.Topics[1], "0x"))
	if err != nil {
		return nil, errors.Wrap(err, "invalid from address")
	}
	to, err := sdk.AccAddressFromHexUnsafe(strings.TrimPrefix(tl.Topics[2], "0x"))
	if err != nil {
		return nil, errors.Wrap(err, "invalid to address")
	}
	/*
		tl.Topics[1] = strings.TrimPrefix(strings.TrimPrefix(tl.Topics[1], "0x"), "000000000000000000000000")
		from, err := evmtypes.ContractAddressFromString(ac, tl.Topics[1])
		if err != nil {
			return nil, errors.Wrap(err, "invalid from address")
		}
		tl.Topics[2] = strings.TrimPrefix(strings.TrimPrefix(tl.Topics[2], "0x"), "000000000000000000000000")
		to, err := evmtypes.ContractAddressFromString(ac, tl.Topics[2])
		if err != nil {
			return nil, errors.Wrap(err, "invalid to address")
		}
	*/

	return &ParsedTransfer{
		Address: addr,
		From:    from,
		To:      to,
		TokenId: tl.Topics[3],
	}, nil
}

func (pt ParsedTransfer) GetAction() NftAction {
	emptyAddr, _ := sdk.AccAddressFromHexUnsafe("0000000000000000000000000000000000000000000000000000000000000000")
	if pt.From.Equals(emptyAddr) && !pt.To.Equals(emptyAddr) {
		return NftActionMint
	}
	if !pt.From.Equals(emptyAddr) && pt.To.Equals(emptyAddr) {
		return NftActionBurn
	}
	return NftActionTransfer
}
