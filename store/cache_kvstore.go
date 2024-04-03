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
	cache *lru.ARCCache
}

func (c CosmosKVStore) GetStoreType() types.StoreType {
	//TODO implement me
	panic("implement me")
}

func (c CosmosKVStore) CacheWrap() types.CacheWrap {
	//TODO implement me
	panic("implement me")
}

func (c CosmosKVStore) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	//TODO implement me
	panic("implement me")
}

func (c CosmosKVStore) Get(key []byte) []byte {
	//TODO implement me
	panic("implement me")
}

func (c CosmosKVStore) Has(key []byte) bool {
	//TODO implement me
	panic("implement me")
}

func (c CosmosKVStore) Set(key, value []byte) {
	//TODO implement me
	panic("implement me")
}

func (c CosmosKVStore) Delete(key []byte) {
	//TODO implement me
	panic("implement me")
}

func (c CosmosKVStore) Iterator(start, end []byte) types.Iterator {
	//TODO implement me
	panic("implement me")
}

func (c CosmosKVStore) ReverseIterator(start, end []byte) types.Iterator {
	//TODO implement me
	panic("implement me")
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
