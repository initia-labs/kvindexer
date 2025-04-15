// GasKVStore is copied from github.com/cosmos/cosmos-sdk/store/gaskv
// and modified to use corestore.KVStore instead of types.KVStore
package store

import (
	corestore "cosmossdk.io/core/store"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ corestore.KVStore = &GasKVStore{}

// GasKVStore is a wrapper of corestore.KVStore that consumes gas for each operation
type GasKVStore struct {
	parent corestore.KVStore

	gasMeter  storetypes.GasMeter
	gasConfig storetypes.GasConfig
}

// NewGasKVStore creates a new GasKVStore
func NewGasKVStore(ctx sdk.Context, store corestore.KVStore) corestore.KVStore {
	return &GasKVStore{
		parent:    store,
		gasMeter:  ctx.GasMeter(),
		gasConfig: ctx.KVGasConfig(),
	}
}

// Implements KVStore.
func (gs *GasKVStore) Get(key []byte) (value []byte, err error) {
	gs.gasMeter.ConsumeGas(gs.gasConfig.ReadCostFlat, storetypes.GasReadCostFlatDesc)
	value, err = gs.parent.Get(key)
	if err != nil {
		return nil, err
	}

	// TODO overflow-safe math?
	gs.gasMeter.ConsumeGas(gs.gasConfig.ReadCostPerByte*storetypes.Gas(len(key)), storetypes.GasReadPerByteDesc)
	gs.gasMeter.ConsumeGas(gs.gasConfig.ReadCostPerByte*storetypes.Gas(len(value)), storetypes.GasReadPerByteDesc)

	return value, nil
}

// Implements KVStore.
func (gs *GasKVStore) Set(key, value []byte) error {
	storetypes.AssertValidKey(key)
	storetypes.AssertValidValue(value)
	gs.gasMeter.ConsumeGas(gs.gasConfig.WriteCostFlat, storetypes.GasWriteCostFlatDesc)
	// TODO overflow-safe math?
	gs.gasMeter.ConsumeGas(gs.gasConfig.WriteCostPerByte*storetypes.Gas(len(key)), storetypes.GasWritePerByteDesc)
	gs.gasMeter.ConsumeGas(gs.gasConfig.WriteCostPerByte*storetypes.Gas(len(value)), storetypes.GasWritePerByteDesc)
	return gs.parent.Set(key, value)
}

// Implements KVStore.
func (gs *GasKVStore) Has(key []byte) (bool, error) {
	gs.gasMeter.ConsumeGas(gs.gasConfig.HasCost, storetypes.GasHasDesc)
	return gs.parent.Has(key)
}

// Implements KVStore.
func (gs *GasKVStore) Delete(key []byte) error {
	// charge gas to prevent certain attack vectors even though space is being freed
	gs.gasMeter.ConsumeGas(gs.gasConfig.DeleteCost, storetypes.GasDeleteDesc)
	return gs.parent.Delete(key)
}

// Iterator implements store.KVStore.
func (gs *GasKVStore) Iterator(start []byte, end []byte) (corestore.Iterator, error) {
	return gs.iterator(start, end, true)
}

// ReverseIterator implements store.KVStore.
func (gs *GasKVStore) ReverseIterator(start []byte, end []byte) (corestore.Iterator, error) {
	return gs.iterator(start, end, false)
}

func (gs *GasKVStore) iterator(start, end []byte, ascending bool) (corestore.Iterator, error) {
	var parent corestore.Iterator
	var err error
	if ascending {
		parent, err = gs.parent.Iterator(start, end)
	} else {
		parent, err = gs.parent.ReverseIterator(start, end)
	}
	if err != nil {
		return nil, err
	}

	gi := newGasIterator(gs.gasMeter, gs.gasConfig, parent)
	gi.(*gasIterator).consumeSeekGas()

	return gi, nil
}

var _ corestore.Iterator = &gasIterator{}

type gasIterator struct {
	gasMeter  storetypes.GasMeter
	gasConfig storetypes.GasConfig
	parent    corestore.Iterator
}

func newGasIterator(gasMeter storetypes.GasMeter, gasConfig storetypes.GasConfig, parent corestore.Iterator) corestore.Iterator {
	return &gasIterator{
		gasMeter:  gasMeter,
		gasConfig: gasConfig,
		parent:    parent,
	}
}

// Implements Iterator.
func (gi *gasIterator) Domain() (start, end []byte) {
	return gi.parent.Domain()
}

// Implements Iterator.
func (gi *gasIterator) Valid() bool {
	return gi.parent.Valid()
}

// Next implements the Iterator interface. It seeks to the next key/value pair
// in the iterator. It incurs a flat gas cost for seeking and a variable gas
// cost based on the current value's length if the iterator is valid.
func (gi *gasIterator) Next() {
	gi.consumeSeekGas()
	gi.parent.Next()
}

// Key implements the Iterator interface. It returns the current key and it does
// not incur any gas cost.
func (gi *gasIterator) Key() (key []byte) {
	key = gi.parent.Key()
	return key
}

// Value implements the Iterator interface. It returns the current value and it
// does not incur any gas cost.
func (gi *gasIterator) Value() (value []byte) {
	value = gi.parent.Value()
	return value
}

// Implements Iterator.
func (gi *gasIterator) Close() error {
	return gi.parent.Close()
}

// Error delegates the Error call to the parent iterator.
func (gi *gasIterator) Error() error {
	return gi.parent.Error()
}

// consumeSeekGas consumes on each iteration step a flat gas cost and a variable gas cost
// based on the current value's length.
func (gi *gasIterator) consumeSeekGas() {
	if gi.Valid() {
		key := gi.Key()
		value := gi.Value()

		gi.gasMeter.ConsumeGas(gi.gasConfig.ReadCostPerByte*storetypes.Gas(len(key)), storetypes.GasValuePerByteDesc)
		gi.gasMeter.ConsumeGas(gi.gasConfig.ReadCostPerByte*storetypes.Gas(len(value)), storetypes.GasValuePerByteDesc)
	}
	gi.gasMeter.ConsumeGas(gi.gasConfig.IterNextCostFlat, storetypes.GasIterNextCostFlatDesc)
}
