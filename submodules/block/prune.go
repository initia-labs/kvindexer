package block

import (
	"context"

	"cosmossdk.io/collections"
)

func (bs BlockSubmodule) prune(ctx context.Context, minHeight int64) error {
	rn := new(collections.Range[int64]).StartInclusive(1).EndInclusive(minHeight)
	return bs.blockByHeight.Clear(ctx, rn)
}
