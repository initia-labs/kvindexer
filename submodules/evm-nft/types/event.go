package types

import (
	"encoding/json"
	"errors"
	fmt "fmt"

	"cosmossdk.io/core/address"
	"github.com/spf13/cast"

	sdk "github.com/cosmos/cosmos-sdk/types"

	evmtypes "github.com/initia-labs/minievm/x/evm/types"
)

type TransferLog struct {
	Address string   `json:"address"`
	Topics  []string `json:"topics"`
}

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
		return nil, errors.New("not transfer log")
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

type MintEvent struct {
	Action          string         `json:"action"`
	ContractAddress sdk.AccAddress `json:"_contract_address"`
	Minter          sdk.AccAddress `json:"minter"`
	Owner           sdk.AccAddress `json:"owner"`
	TokenId         string         `json:"token_id"`
	MsgIdx          uint64         `json:"msg_index"`
}

func getUint64FromMap(src EventWithAttributeMap, key string) (uint64, error) {
	val, err := cast.ToUint64E(src.AttributesMap[key])
	if err != nil {
		return 0, fmt.Errorf("%s is invalid", key)
	}
	return val, nil
}

func getStringFromMap(src EventWithAttributeMap, key string) (string, error) {
	val, found := src.AttributesMap[key]
	if !found {
		return "", fmt.Errorf("%s is invalid", key)
	}
	return val, nil
}

func getSdkAddressFromMap(src EventWithAttributeMap, key string) (sdk.AccAddress, error) {
	addr, err := sdk.AccAddressFromBech32(src.AttributesMap[key])
	if err != nil {
		return nil, fmt.Errorf("%s is invalid", key)
	}
	return addr, nil
}

func (event *MintEvent) Parse(src EventWithAttributeMap) (err error) {
	if event.Action, err = getStringFromMap(src, "action"); err != nil {
		return err
	}
	if event.ContractAddress, err = getSdkAddressFromMap(src, "_contract_address"); err != nil {
		return err
	}
	if event.Minter, err = getSdkAddressFromMap(src, "minter"); err != nil {
		return err
	}
	if event.Owner, err = getSdkAddressFromMap(src, "owner"); err != nil {
		return err
	}
	if event.TokenId, err = getStringFromMap(src, "token_id"); err != nil {
		return err
	}
	if event.MsgIdx, err = getUint64FromMap(src, "msg_index"); err != nil {
		return err
	}
	return nil
}

type TransferOrSendEvent struct {
	Action          string         `json:"action"`
	ContractAddress sdk.AccAddress `json:"_contract_address"`
	Recipient       sdk.AccAddress `json:"recipient"`
	Sender          sdk.AccAddress `json:"sender"`
	TokenId         string         `json:"token_id"`
	MsgIdx          uint64         `json:"msg_index"`
}

func (event *TransferOrSendEvent) Parse(src EventWithAttributeMap) (err error) {
	if event.Action, err = getStringFromMap(src, "action"); err != nil {
		return err
	}
	if event.ContractAddress, err = getSdkAddressFromMap(src, "_contract_address"); err != nil {
		return err
	}
	if event.Recipient, err = getSdkAddressFromMap(src, "recipient"); err != nil {
		return err
	}
	if event.Sender, err = getSdkAddressFromMap(src, "sender"); err != nil {
		return err
	}
	if event.TokenId, err = getStringFromMap(src, "token_id"); err != nil {
		return err
	}
	if event.MsgIdx, err = getUint64FromMap(src, "msg_index"); err != nil {
		return err
	}
	return nil
}

type BurnEvent struct {
	Action          string         `json:"action"`
	ContractAddress sdk.AccAddress `json:"_contract_address"`
	Sender          sdk.AccAddress `json:"sender"`
	TokenId         string         `json:"token_id"`
	MsgIdx          uint64         `json:"msg_index"`
}

func (event *BurnEvent) Parse(src EventWithAttributeMap) (err error) {
	if event.Action, err = getStringFromMap(src, "action"); err != nil {
		return err
	}
	if event.ContractAddress, err = getSdkAddressFromMap(src, "_contract_address"); err != nil {
		return err
	}
	if event.Sender, err = getSdkAddressFromMap(src, "sender"); err != nil {
		return err
	}
	if event.TokenId, err = getStringFromMap(src, "token_id"); err != nil {
		return err
	}
	if event.MsgIdx, err = getUint64FromMap(src, "msg_index"); err != nil {
		return err
	}
	return nil
}
