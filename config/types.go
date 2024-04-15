package config

import (
	"github.com/spf13/viper"
)

type IndexerConfig struct {
	Enable        bool         `mapstructure:"indexer.enable"`
	CacheSize     uint         `mapstructure:"indexer.cache-size"`
	L1ChainId     string       `mapstructure:"indexer.l1-chain-id"`
	BackendConfig *viper.Viper `mapstructure:"indexer.backend"`
}

const DefaultConfigTemplate = `
###############################################################################
###                              Indexer                                   ###
###############################################################################

[indexer]

# Enable defines whether the indexer is enabled.
enable = {{ .IndexerConfig.Enable }}

# CacheSize defines the size of the cache.
cache-size = {{ .IndexerConfig.CacheSize }}

# l1-chain-id defines the chain id of the l1 chain.
l1-chain-id = {{ .IndexerConfig.L1ChainID }}

# Backend defines the type of the backend store and its options.
# It should have a key-value pair named 'type', and the value should exist in store supported by cosmos-db.
# supported type: "goleveldb" only in current
[indexer.backend]
{{ range $key, $value := .IndexerConfig.BackendConfig.AllSettings }}{{ printf "%s = \"%v\"\n" $key $value }}{{end}}
`
