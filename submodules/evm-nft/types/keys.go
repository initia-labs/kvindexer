package types

const (
	// SubmoduleName is the name of the submodule
	SubmoduleName = "evm-nft"

	// Version is the current version of the submodule
	Version = "v0.1.9"
)

// store prefixes
const (
	CollectionsPrefix      = 0x10
	CollectionOwnersPrefix = 0x20
	CollectionNamesPrefix  = 0x21

	TokensPrefix      = 0x30
	TokenOwnersPrefix = 0x40

	MigrationPrefix = 0xff
)
