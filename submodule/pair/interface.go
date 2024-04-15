package pair

import (
	"context"
)

func GetPair(ctx context.Context, isFungible bool, l2key string) (string, error) {
	if isFungible {
		return fungiblePairsMap.Get(ctx, l2key)
	}
	return nonFungiblePairsMap.Get(ctx, l2key)
}
