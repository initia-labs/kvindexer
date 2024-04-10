package keeper

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
)

type VMKeeper struct {
	wasmkeeper.Keeper
}

