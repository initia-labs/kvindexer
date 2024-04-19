package collection

import (
	"errors"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/codec"
)

// NewPrefix returns a new prefix
func NewPrefix[T interface{ int | string | []byte }](submodule string, identifier T) collections.Prefix {
	return append([]byte(submodule), []byte(collections.NewPrefix(identifier))...)
}

// AddMap adds a map collection to the keeper
func AddMap[K, V any](k IndexerKeeper, prefix collections.Prefix, name string, kc codec.KeyCodec[K], vc codec.ValueCodec[V]) (*collections.Map[K, V], error) {
	if k.IsSealed() {
		return nil, errors.New("cannot add collection to sealed keeper")
	}
	m := collections.NewMap(k.GetSchemaBuilder(), prefix, name, kc, vc)
	return &m, nil
}

// AddSequence adds a sequence collection to the keeper
func AddSequence(k IndexerKeeper, prefix collections.Prefix, name string) (*collections.Sequence, error) {
	if k.IsSealed() {
		return nil, errors.New("cannot add collection to sealed keeper")
	}
	seq := collections.NewSequence(k.GetSchemaBuilder(), prefix, name)
	return &seq, nil
}

// AddKeySet adds a key set collection to the keeper
func AddKeySet[K any](k IndexerKeeper, prefix collections.Prefix, name string, kc codec.KeyCodec[K]) (*collections.KeySet[K], error) {
	if k.IsSealed() {
		return nil, errors.New("cannot add collection to sealed keeper")
	}
	ks := collections.NewKeySet(k.GetSchemaBuilder(), prefix, name, kc)
	return &ks, nil
}

// AddValueSet adds a value set collection to the keeper
func AddIndexedMap[PrimaryKey, Value any, Idx collections.Indexes[PrimaryKey, Value]](k IndexerKeeper, prefix collections.Prefix, name string, pkCodec codec.KeyCodec[PrimaryKey], valueCodec codec.ValueCodec[Value], indices Idx) (*collections.IndexedMap[PrimaryKey, Value, Idx], error) {
	if k.IsSealed() {
		return nil, errors.New("cannot add collection to sealed keeper")
	}
	im := collections.NewIndexedMap(k.GetSchemaBuilder(), prefix, name, pkCodec, valueCodec, indices)
	return im, nil
}

// AddItem adds an item collection to the keeper
func AddItem[V any](k IndexerKeeper, prefix collections.Prefix, name string, vc codec.ValueCodec[V]) (*collections.Item[V], error) {
	if k.IsSealed() {
		return nil, errors.New("cannot add collection to sealed keeper")
	}
	item := collections.NewItem(k.GetSchemaBuilder(), prefix, name, vc)
	return &item, nil
}
