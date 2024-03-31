//go:build !(vm_wasm || vm_eth)

package keeper

import (
	movekeeper "github.com/initia-labs/initia/x/move/keeper"
)

type VMKeeper struct {
	movekeeper.Keeper
}
