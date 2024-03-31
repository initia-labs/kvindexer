package nft

import (
	"context"

	"github.com/initia-labs/kvindexer/config"
	"github.com/initia-labs/kvindexer/module/keeper"
	"github.com/initia-labs/kvindexer/submodule/nft/types"
)

const eventType = "none"

func processEvents(k *keeper.Keeper, ctx context.Context, cfg config.SubmoduleConfig, events []types.EventWithAttributeMap) error {
	panic("not implemented")
}
