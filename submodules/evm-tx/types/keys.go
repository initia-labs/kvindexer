package types

const (
	// SubmoduleName is the name of the submodule
	SubmoduleName = "evm-tx"

	// Version is the current version of the submodule
	Version = "v0.1.3"
)

// store prefixes
const (
	TxsByAccountPrefix            = 0x10
	AccountSequencePrefix         = 0x20
	SequencePrefix                = 0xa0
	TxSequencePrefix              = 0xb0
	TxByHeightPrefix              = 0xc0
	TxsPrefix                     = 0xf0
	SequenceByHeightPrefix        = 0xd0
	AccountSequenceByHeightPrefix = 0xe0
)
