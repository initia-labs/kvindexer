package store

import (
	"io"

	"cosmossdk.io/store/types"
	db "github.com/cosmos/cosmos-db"
	"github.com/patrickmn/go-cache"
)

type KVStore interface {
	Get(key []byte) []byte
	Has(key []byte) bool
	Set(key, value []byte)
	Delete(key []byte)
}

type CosmosKVStore struct {
	store types.CacheKVStore
	cache *cache.Cache
}

func NewCosmosKVStore(store types.CacheKVStore) types.CacheKVStore {
	return CosmosKVStore{
		store: store,
		cache: cache.New(cache.NoExpiration, cache.NoExpiration),
	}
}

// CacheWrap implements types.CacheKVStore.
func (c CosmosKVStore) CacheWrap() types.CacheWrap {
	panic("unimplemented")
}

// CacheWrapWithTrace implements types.CacheKVStore.
func (c CosmosKVStore) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	panic("unimplemented")
}

// GetStoreType implements types.CacheKVStore.
func (c CosmosKVStore) GetStoreType() types.StoreType {
	panic("unimplemented")
}

// Iterator implements types.CacheKVStore.
func (c CosmosKVStore) Iterator(start []byte, end []byte) db.Iterator {
	panic("unimplemented")
}

// ReverseIterator implements types.CacheKVStore.
func (c CosmosKVStore) ReverseIterator(start []byte, end []byte) db.Iterator {
	panic("unimplemented")
}

func (c CosmosKVStore) Get(key []byte) []byte {
	if value, found := c.cache.Get(string(key)); found {
		return value.([]byte)
	}
	val := c.store.Get(key)
	c.cache.Set(string(key), val, cache.NoExpiration)
	return val
}

func (c CosmosKVStore) Has(key []byte) bool {
	value := c.store.Get(key)
	return value != nil
}

func (c CosmosKVStore) Set(key, value []byte) {
	c.cache.Set(string(key), value, cache.NoExpiration)
	c.store.Set(key, value)
}

func (c CosmosKVStore) Delete(key []byte) {
	c.cache.Delete(string(key))
	c.store.Delete(key)
}

func (c CosmosKVStore) Write() {
	c.store.Write()
	panic("unimplemented")
}
