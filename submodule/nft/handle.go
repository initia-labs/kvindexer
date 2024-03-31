//go:build !(vm_move || vm_wasm || vm_evm)

package nft

import (
	"context"

	"github.com/initia-labs/indexer/v2/config"
	"github.com/initia-labs/indexer/v2/module/keeper"
	"github.com/initia-labs/indexer/v2/submodule/nft/types"
)

const eventType = "none"

func processEvents(k *keeper.Keeper, ctx context.Context, cfg config.SubmoduleConfig, events []types.EventWithAttributeMap) error {
	panic("not implemented")
}
