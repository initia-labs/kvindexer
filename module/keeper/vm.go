package keeper

import (
	evmkeeper "github.com/initia-labs/minievm/x/evm/keeper"
)

type VMKeeper struct {
	*evmkeeper.Keeper
}

func (k VMKeeper) GetVMType() string {
	return "evm"
}
