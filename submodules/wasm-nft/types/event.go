package types

import (
	fmt "fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/initia-labs/kvindexer/submodules/wasm-nft/util"
	"github.com/spf13/cast"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MintEvent struct {
	Action          string         `json:"action"`
	ContractAddress sdk.AccAddress `json:"_contract_address"`
	Minter          sdk.AccAddress `json:"minter"`
	Owner           sdk.AccAddress `json:"owner"`
	TokenId         string         `json:"token_id"`
	MsgIdx          uint64         `json:"msg_index"`
}

func getUint64FromMap(src abci.Event, key string) (uint64, error) {
	valuestr, err := util.GetAttributeValue(src, key)
	if err != nil {
		return 0, err
	}
	val, err := cast.ToUint64E(valuestr)
	if err != nil {
		return 0, fmt.Errorf("%s is invalid", key)
	}
	return val, nil
}

func getStringFromMap(src abci.Event, key string) (string, error) {
	return util.GetAttributeValue(src, key)
}

func getSdkAddressFromMap(src abci.Event, key string) (sdk.AccAddress, error) {
	valuestr, err := util.GetAttributeValue(src, key)
	if err != nil {
		return nil, err
	}
	addr, err := sdk.AccAddressFromBech32(valuestr)
	if err != nil {
		return nil, fmt.Errorf("%s is invalid", key)
	}
	return addr, nil
}

func (event *MintEvent) Parse(src abci.Event) (err error) {
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

func (event *TransferOrSendEvent) Parse(src abci.Event) (err error) {
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

func (event *BurnEvent) Parse(src abci.Event) (err error) {
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
