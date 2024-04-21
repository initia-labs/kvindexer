package collection

import (
	"cosmossdk.io/collections"
)

type IndexerKeeper interface {
	IsSealed() bool
	GetSchemaBuilder() *collections.SchemaBuilder
}
