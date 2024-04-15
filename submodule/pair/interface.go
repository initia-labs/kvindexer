package pair

import (
	"context"

	"cosmossdk.io/collections"
	cosmoserr "cosmossdk.io/errors"
)

func GetPair(ctx context.Context, isFungible bool, l2key string) (string, error) {
	target := nonFungiblePairsMap
	if isFungible {
		target = fungiblePairsMap
	}
	return target.Get(ctx, l2key)
}
func SetPair(ctx context.Context, overwrite, isFungible bool, l2key, l1key string) error {
	target := nonFungiblePairsMap
	if isFungible {
		target = fungiblePairsMap
	}

	prev, err := target.Get(ctx, l2key)
	if err != nil {
		if cosmoserr.IsOf(err, collections.ErrNotFound) {
			return target.Set(ctx, l2key, l1key)
		}
		return err
	}
	if !overwrite {
		return nil
	}
	if prev == l1key {
		return nil
	}
	return target.Set(ctx, l2key, l1key)
}

func SetPair(ctx context.Context, isFungible bool, l2key, l1key string) error {
	if isFungible {
		return fungiblePairsMap.Set(ctx, l2key, l1key)
	}
	return nonFungiblePairsMap.Set(ctx, l2key, l1key)
}
