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
	//TODO implement me
	panic("implement me - GET")
}

func (c CacheStore) Has(key []byte) (bool, error) {
	//TODO implement me
	panic("implement me- HAS")
}

func (c CacheStore) Set(key, value []byte) error {
	//TODO implement me
	panic("implement me - SET")
}

func (c CacheStore) Delete(key []byte) error {
	//TODO implement me
	panic("implement me - DELETE")
}

func (c CacheStore) Iterator(start, end []byte) (types.Iterator, error) {
	//TODO implement me
	panic("implement me - ITER")
}

func (c CacheStore) ReverseIterator(start, end []byte) (types.Iterator, error) {
	//TODO implement me
	panic("implement me - REV")
}
