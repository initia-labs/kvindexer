package types

import (
	fmt "fmt"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/spf13/cast"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	eventKeyAction          = "action"
	eventKeyContractAddress = "_contract_address"
	eventKeyMinter          = "minter"
	eventKeyOwner           = "owner"
	eventKeyRecipient       = "recipient"
	eventKeySender          = "sender"
	eventKeyTokenId         = "token_id"
	eventKeyMsgIndex        = "msg_index"
)

type MintEvent struct {
	Action          string         `json:"action"`
	ContractAddress sdk.AccAddress `json:"_contract_address"`
	Minter          sdk.AccAddress `json:"minter"`
	Owner           sdk.AccAddress `json:"owner"`
	TokenId         string         `json:"token_id"`
	MsgIdx          uint64         `json:"msg_index"`
}

//nolint:dupl
func (event *MintEvent) Parse(src abci.Event) (err error) {
	for _, attr := range src.Attributes {
		switch attr.Key {
		case eventKeyAction:
			event.Action = attr.Value
		case eventKeyContractAddress:
			event.ContractAddress, err = sdk.AccAddressFromBech32(attr.Value)
			if err != nil {
				return fmt.Errorf("_contract_address is invalid")
			}
		case eventKeyMinter:
			event.Minter, err = sdk.AccAddressFromBech32(attr.Value)
			if err != nil {
				return fmt.Errorf("minter is invalid")
			}
		case eventKeyOwner:
			event.Owner, err = sdk.AccAddressFromBech32(attr.Value)
			if err != nil {
				return fmt.Errorf("owner is invalid")
			}
		case eventKeyTokenId:
			event.TokenId = attr.Value
		case eventKeyMsgIndex:
			event.MsgIdx, err = cast.ToUint64E(attr.Value)
			if err != nil {
				return fmt.Errorf("msg_index is invalid")
			}
		}
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

//nolint:dupl
func (event *TransferOrSendEvent) Parse(src abci.Event) (err error) {
	for _, attr := range src.Attributes {
		switch attr.Key {
		case eventKeyAction:
			event.Action = attr.Value
		case eventKeyContractAddress:
			event.ContractAddress, err = sdk.AccAddressFromBech32(attr.Value)
			if err != nil {
				return fmt.Errorf("_contract_address is invalid")
			}
		case eventKeyRecipient:
			event.Recipient, err = sdk.AccAddressFromBech32(attr.Value)
			if err != nil {
				return fmt.Errorf("recipient is invalid")
			}
		case eventKeySender:
			event.Sender, err = sdk.AccAddressFromBech32(attr.Value)
			if err != nil {
				return fmt.Errorf("sender is invalid")
			}
		case eventKeyTokenId:
			event.TokenId = attr.Value
		case eventKeyMsgIndex:
			event.MsgIdx, err = cast.ToUint64E(attr.Value)
			if err != nil {
				return fmt.Errorf("msg_index is invalid")
			}
		}
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
	for _, attr := range src.Attributes {
		switch attr.Key {
		case eventKeyAction:
			event.Action = attr.Value
		case eventKeyContractAddress:
			event.ContractAddress, err = sdk.AccAddressFromBech32(attr.Value)
			if err != nil {
				return fmt.Errorf("_contract_address is invalid")
			}
		case eventKeySender:
			event.Sender, err = sdk.AccAddressFromBech32(attr.Value)
			if err != nil {
				return fmt.Errorf("sender is invalid")
			}
		case eventKeyTokenId:
			event.TokenId = attr.Value
		case eventKeyMsgIndex:
			event.MsgIdx, err = cast.ToUint64E(attr.Value)
			if err != nil {
				return fmt.Errorf("msg_index is invalid")
			}
		}
	}
	return nil
}
