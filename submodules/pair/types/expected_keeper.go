package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
)

type ChannelKeeper interface {
	GetAllChannels(ctx sdk.Context) (channels []channeltypes.IdentifiedChannel)
	GetChannelClientState(ctx sdk.Context, portID, channelID string) (string, ibcexported.ClientState, error)
}

type TransferKeeper interface {
	GetAllDenomTraces(ctx sdk.Context) transfertypes.Traces
}
