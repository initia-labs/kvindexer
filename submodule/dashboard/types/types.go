package types

import context "context"

type Uint64Map interface {
	Get(ctx context.Context, key string) (value uint64, err error)
	Set(ctx context.Context, key string, value uint64) error
}
