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

func SetPair(ctx context.Context, isFungible bool, l2key, l1key string) error {
	if !enabled {
		return fmt.Errorf("cannot set: %s is disabled", submoduleName)
	}
	if isFungible {
		return fungiblePairsMap.Set(ctx, l2key, l1key)
	}
	return nonFungiblePairsMap.Set(ctx, l2key, l1key)
}
