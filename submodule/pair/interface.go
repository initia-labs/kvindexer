package pair

import (
	"context"
	"fmt"
)

func GetPair(ctx context.Context, isFungible bool, l2key string) (string, error) {
	if !enabled {
		return "", fmt.Errorf("cannot query: %s is disabled", submoduleName)
	}
	if isFungible {
		return fungiblepairsMap.Get(ctx, l2key)
	}
	return nonFungiblepairsMap.Get(ctx, l2key)
}
