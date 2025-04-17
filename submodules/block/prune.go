package block

import (
	"context"

	"cosmossdk.io/collections"
)

func (sub BlockSubmodule) prune(ctx context.Context, minHeight int64) error {
	rn := new(collections.Range[int64]).StartInclusive(1).EndInclusive(minHeight)
	return sub.blockByHeight.Clear(ctx, rn)
}
