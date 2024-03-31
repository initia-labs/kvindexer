package types

const (
	// ModuleName is the name of the module
	ModuleName = "indexer"

	// StoreKey is the string store representation
	// CAUTION: DO NOT STORE CONSENSUS STATE WITH THIS KEY
	StoreKey = ModuleName

	// TStoreKey is the string transient store representation
	TStoreKey = "transient_" + ModuleName

	// QuerierRoute is the querier route for the move module
	QuerierRoute = ModuleName

	// No Router Key for this module
)
