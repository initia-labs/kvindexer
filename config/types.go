package config

import (
	"github.com/spf13/viper"
)

// IndexerConfig defines the configuration options for the kvindexer.
type IndexerConfig struct {
	// Enable defines whether the kvindexer is enabled.
	Enable bool `mapstructure:"indexer.enable"`
	// CacheCapacity defines the size of the cache used by the kvindexer. (unit: MiB)
	CacheCapacity int `mapstructure:"indexer.cache-capacity"`
	// Backend defines the type of the backend store and its options.
	//  It should have a key-value pair named 'type', and the value should exist in store supported by cosmos-db.
	// Recommend to use default value unless you know about backend db storage.
	// NOTE: "goleveldb" is the only supported type in the current version.
	BackendConfig *viper.Viper `mapstructure:"indexer.backend"`
}

const DefaultConfigTemplate = `
###############################################################################
###                              Indexer                                   ###
###############################################################################

[indexer]

# Enable defines whether the indexer is enabled.
enable = {{ .IndexerConfig.Enable }}

# CacheCapacity defines the size of the cache. (unit: MiB)
cache-capacity = {{ .IndexerConfig.CacheCapacity }}

# Backend defines the type of the backend store and its options.
# It should have a key-value pair named 'type', and the value should exist in store supported by cosmos-db.
# Recommend to use default value unless you know about backend db storage.
# supported type: "goleveldb" only in current
[indexer.backend]
{{ range $key, $value := .IndexerConfig.BackendConfig.AllSettings }}{{ printf "%s = \"%v\"\n" $key $value }}{{end}}
`
