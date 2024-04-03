package store

import (
	"cosmossdk.io/store/cachekv"
	"cosmossdk.io/store/types"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"io"
)

type KVStore interface {
	Get(key []byte) []byte
	Has(key []byte) bool
	Set(key, value []byte)
	Delete(key []byte)
}

type CosmosKVStore struct {
	store types.CacheKVStore
	// ARCCache is a thread-safe fixed size Adaptive Replacement Cache (ARC).
	// ARC is an enhancement over the standard LRU cache in that tracks both
	// frequency and recency of use.
	cache *lru.ARCCache
}

func (c CosmosKVStore) GetStoreType() types.StoreType {
	panic("not implemented")
}

func (c CosmosKVStore) CacheWrap() types.CacheWrap {
	return cachekv.NewStore(c)
}

func (c CosmosKVStore) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	panic("not implemented")
}

func (c CosmosKVStore) Get(key []byte) []byte {
	//TODO implement me
	panic("implement me")
}

func (c CosmosKVStore) Has(key []byte) bool {
	types.AssertValidKey(key)
	_, ok := c.cache.Get(key)
	return ok
}

func (c CosmosKVStore) Set(key, value []byte) {
	//TODO implement me
	panic("implement me")
}

func (c CosmosKVStore) Delete(key []byte) {
	c.cache.Remove(key)
	c.store.Delete(key)
}

func (c CosmosKVStore) Iterator(start, end []byte) types.Iterator {
	return c.store.Iterator(start, end)
}

func (c CosmosKVStore) ReverseIterator(start, end []byte) types.Iterator {
	return c.store.ReverseIterator(start, end)
}

func (c CosmosKVStore) Write() {
	//TODO implement me
	panic("implement me")
}

func NewStore(store types.KVStore, size uint) types.CacheKVStore {
	cache, err := lru.NewARC(int(size))
	if err != nil {
		panic(fmt.Errorf("failed to create KVStore cache: %s", err))
	}

	return CosmosKVStore{
		store: cachekv.NewStore(store),
		cache: cache,
	}
}
