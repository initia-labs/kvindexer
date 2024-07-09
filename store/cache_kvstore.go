package store

import (
	"context"

	"cosmossdk.io/errors"
	cachekv "cosmossdk.io/store/cachekv"
	"cosmossdk.io/store/types"
	bigcache "github.com/allegro/bigcache/v3"
)

type CacheStore struct {
	store types.CacheKVStore
	cache *bigcache.BigCache
}

func NewCacheStore(store types.KVStore, capacity int) *CacheStore {
	// default with no eviction and custom hard max cache capacity
	cacheCfg := bigcache.DefaultConfig(0)
	cacheCfg.Verbose = false
	cacheCfg.HardMaxCacheSize = capacity

	cache, err := bigcache.New(context.Background(), cacheCfg)
	if err != nil {
		panic(err)
	}

	return &CacheStore{
		store: cachekv.NewStore(store),
		cache: cache,
	}
}

func (c CacheStore) Get(key []byte) ([]byte, error) {
	types.AssertValidKey(key)

	v, err := c.cache.Get(string(key))
	// cache hit
	if err == nil {
		return v, nil
	}

	// get from store and write to cache
	value := c.store.Get(key)
	err = c.cache.Set(string(key), value)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set cache")
	}

	return value, nil
}

func (c CacheStore) Has(key []byte) (bool, error) {
	_, err := c.cache.Get(string(key))
	return err == nil, err
}

func (c CacheStore) Set(key, value []byte) error {
	types.AssertValidKey(key)
	types.AssertValidValue(value)

	err := c.cache.Set(string(key), value)
	if err != nil {
		return errors.Wrap(err, "failed to set cache")
	}
	c.store.Set(key, value)

	return nil
}

func (c CacheStore) Delete(key []byte) error {
	err := c.cache.Delete(string(key))
	if err != nil && errors.IsOf(err, bigcache.ErrEntryNotFound) {
		return errors.Wrap(err, "failed to delete cache")
	}
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
