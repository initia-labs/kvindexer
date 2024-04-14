package keeper

import (
	movekeeper "github.com/initia-labs/initia/x/move/keeper"
)

type VMKeeper struct {
	movekeeper.Keeper
}

func (k VMKeeper) GetVMType() string {
	return "move"
}
