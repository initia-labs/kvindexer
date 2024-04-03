package store

import (
	"cosmossdk.io/store/types"
)

type KVStore interface {
	Get(key []byte) []byte
	Has(key []byte) bool
	Set(key, value []byte)
	Delete(key []byte)
}

type CosmosKVStore struct {
	store types.CacheKVStore
}

func NewCosmosKVStore(store types.CacheKVStore) *CosmosKVStore {
	return &CosmosKVStore{store: store}
}

func (c *CosmosKVStore) Get(key []byte) []byte {
	return c.store.Get(key)
}

func (c *CosmosKVStore) Has(key []byte) bool {
	value := c.store.Get(key)
	return value != nil
}

func (c *CosmosKVStore) Set(key, value []byte) {
	c.store.Set(key, value)
}

func (c *CosmosKVStore) Delete(key []byte) {
	c.store.Delete(key)
}
