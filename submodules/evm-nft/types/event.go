package types

import (
	"encoding/json"
	"math/big"
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

func (tl TransferLog) IsErc721Transfer() bool {
	return (len(tl.Topics) == 4) && (tl.Topics[0] == "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef") && (tl.Data == "0x")
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

	from, err := sdk.AccAddressFromHexUnsafe(strings.TrimPrefix(strings.TrimPrefix(tl.Topics[1], "0x"), "000000000000000000000000"))
	//from, err := sdk.AccAddressFromHexUnsafe(strings.TrimPrefix(tl.Topics[1], "0x"))
	if err != nil {
		return nil, errors.Wrap(err, "invalid from address")
	}
	to, err := sdk.AccAddressFromHexUnsafe(strings.TrimPrefix(strings.TrimPrefix(tl.Topics[2], "0x"), "000000000000000000000000"))
	//to, err := sdk.AccAddressFromHexUnsafe(strings.TrimPrefix(tl.Topics[2], "0x"))
	if err != nil {
		return nil, errors.Wrap(err, "invalid to address")
	}
	tokenId, err := convertHexStringToDecString(tl.Topics[3])
	if err != nil {
		return nil, errors.Wrap(err, "invalid token id")
	}

	return &ParsedTransfer{
		Address: addr,
		From:    from,
		To:      to,
		TokenId: tokenId,
	}, nil
}

func (pt ParsedTransfer) GetAction() NftAction {
	emptyAddr, _ := sdk.AccAddressFromHexUnsafe("0000000000000000000000000000000000000000")
	//emptyAddr, _ := sdk.AccAddressFromHexUnsafe("0000000000000000000000000000000000000000000000000000000000000000")
	if pt.From.Equals(emptyAddr) && !pt.To.Equals(emptyAddr) {
		return NftActionMint
	}
	if !pt.From.Equals(emptyAddr) && pt.To.Equals(emptyAddr) {
		return NftActionBurn
	}
	return NftActionTransfer
}

// NOTE: non-hexadecimal input causes unexpected results
func convertHexStringToDecString(hex string) (string, error) {
	hex = strings.TrimPrefix(hex, "0x")
	bi, ok := new(big.Int).SetString(hex, 16)
	if !ok {
		return "", errors.New("failed to convert hex to dec")
	}
	return bi.String(), nil
}
