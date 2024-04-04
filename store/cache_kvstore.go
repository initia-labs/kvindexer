package store

import (
	"cosmossdk.io/store/types"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
)

type CacheStore struct {
	store types.KVStore
	cache *lru.ARCCache
}

func NewCacheStore(store types.KVStore, size uint) *CacheStore {
	cache, err := lru.NewARC(int(size))
	if err != nil {
		panic(fmt.Errorf("failed to create KVStore cache: %s", err))
	}

	return &CacheStore{
		store: store,
		cache: cache,
	}
}

func (c CacheStore) Get(key []byte) ([]byte, error) {
	types.AssertValidKey(key)

	v, ok := c.cache.Get(key)
	if ok {
		// cache hit
		return v.([]byte), nil
	}

	// write to cache
	value := c.store.Get(key)
	c.cache.Add(key, value)

	return value, nil
}

func (c CacheStore) Has(key []byte) (bool, error) {
	_, ok := c.cache.Get(key)
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
	panic("not implemented")
}

func (c CacheStore) ReverseIterator(start, end []byte) (types.Iterator, error) {
	panic("not implemented")
}
