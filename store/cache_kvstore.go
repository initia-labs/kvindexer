package store

import (
	"fmt"

	cachekv "cosmossdk.io/store/cachekv"
	"cosmossdk.io/store/types"
	lru "github.com/hashicorp/golang-lru"
)

type CacheStore struct {
	store types.CacheKVStore
	cache *lru.ARCCache
}

func NewCacheStore(store types.KVStore, size uint) *CacheStore {
	cache, err := lru.NewARC(int(size))
	if err != nil {
		panic(fmt.Errorf("failed to create KVStore cache: %s", err))
	}

	return &CacheStore{
		store: cachekv.NewStore(store),
		cache: cache,
	}
}

func (c CacheStore) Get(key []byte) ([]byte, error) {
	types.AssertValidKey(key)

	v, ok := c.cache.Get(string(key))
	if ok {
		// cache hit
		return v.([]byte), nil
	}

	// write to cache
	value := c.store.Get(key)
	c.cache.Add(string(key), value)

	return value, nil
}

func (c CacheStore) Has(key []byte) (bool, error) {
	_, ok := c.cache.Get(string(key))
	return ok, nil
}

func (c CacheStore) Set(key, value []byte) error {
	types.AssertValidKey(key)
	types.AssertValidValue(value)

	c.cache.Add(string(key), value)
	c.store.Set(key, value)

	return nil
}

func (c CacheStore) Delete(key []byte) error {
	c.cache.Remove(string(key))
	c.store.Delete(key)

	return nil
}

func (c CacheStore) Iterator(start, end []byte) (types.Iterator, error) {
	return c.store.Iterator(start, end), nil
}

func (c CacheStore) ReverseIterator(start, end []byte) (types.Iterator, error) {
	return c.store.ReverseIterator(start, end), nil
}

func (c CacheStore) Write() {
	c.store.Write()
}
