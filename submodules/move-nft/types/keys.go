package types

const (
	// SubmoduleName is the name of the submodule
	SubmoduleName = "move-nft"

	// Version is the current version of the submodule
	Version = "v0.2.0"
)

// store prefixes
const (
	CollectionsPrefix      = 0x10
	CollectionOwnersPrefix = 0x20
	CollectionNamesPrefix  = 0x21

	TokensPrefix            = 0x30
	TokenAddressIndexPrefix = 0x31

	TokenOwnersPrefix = 0x40

	MigrationPrefix = 0xff
)
