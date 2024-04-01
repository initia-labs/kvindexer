package keeper

import (
	"errors"

	"cosmossdk.io/collections"
	col "cosmossdk.io/collections"
	ccdc "cosmossdk.io/collections/codec"
)

// NewPrefix returns a new prefix
func NewPrefix[T interface{ int | string | []byte }](submodule string, identifier T) collections.Prefix {
	return append([]byte(submodule), []byte(collections.NewPrefix(identifier))...)
}

// AddMap adds a map collection to the keeper
func AddMap[K, V any](k *Keeper, prefix col.Prefix, name string, kc ccdc.KeyCodec[K], vc ccdc.ValueCodec[V]) (*col.Map[K, V], error) {
	if k.IsSealed() {
		return nil, errors.New("cannot add collection to sealed keeper")
	}
	m := col.NewMap(k.schemaBuilder, prefix, name, kc, vc)
	return &m, nil
}

// AddSequence adds a sequence collection to the keeper
func AddSequence(k *Keeper, prefix col.Prefix, name string) (*col.Sequence, error) {
	if k.IsSealed() {
		return nil, errors.New("cannot add collection to sealed keeper")
	}
	seq := col.NewSequence(k.schemaBuilder, prefix, name)
	return &seq, nil
}

// AddKeySet adds a key set collection to the keeper
func AddKeySet[K any](k *Keeper, prefix col.Prefix, name string, kc ccdc.KeyCodec[K]) (*col.KeySet[K], error) {
	if k.IsSealed() {
		return nil, errors.New("cannot add collection to sealed keeper")
	}
	ks := col.NewKeySet(k.schemaBuilder, prefix, name, kc)
	return &ks, nil
}

// AddValueSet adds a value set collection to the keeper
func AddIndexedMap[PrimaryKey, Value any, Idx col.Indexes[PrimaryKey, Value]](k *Keeper, prefix col.Prefix, name string, pkCodec ccdc.KeyCodec[PrimaryKey], valueCodec ccdc.ValueCodec[Value], indices Idx) (*col.IndexedMap[PrimaryKey, Value, Idx], error) {
	if k.IsSealed() {
		return nil, errors.New("cannot add collection to sealed keeper")
	}
	im := col.NewIndexedMap(k.schemaBuilder, prefix, name, pkCodec, valueCodec, indices)
	return im, nil
}

// AddItem adds an item collection to the keeper
func AddItem[V any](k *Keeper, prefix col.Prefix, name string, vc ccdc.ValueCodec[V]) (*col.Item[V], error) {
	if k.IsSealed() {
		return nil, errors.New("cannot add collection to sealed keeper")
	}
	item := col.NewItem(k.schemaBuilder, prefix, name, vc)
	return &item, nil
}
