package config

import (
	"github.com/spf13/viper"
)

type IndexerConfig struct {
	Enable        bool         `mapstructure:"indexer.enable"`
	CacheCapacity int          `mapstructure:"indexer.cache-capacity"`
	BackendConfig *viper.Viper `mapstructure:"indexer.backend"`
}

const DefaultConfigTemplate = `
###############################################################################
###                              Indexer                                   ###
###############################################################################

[indexer]

# Enable defines whether the indexer is enabled.
enable = {{ .IndexerConfig.Enable }}

# CacheCapacity defines the size of the cache. (unit: bytes)
cache-capacity = {{ .IndexerConfig.CacheCapacity }}

# Backend defines the type of the backend store and its options.
# It should have a key-value pair named 'type', and the value should exist in store supported by cosmos-db.
# supported type: "goleveldb" only in current
[indexer.backend]
{{ range $key, $value := .IndexerConfig.BackendConfig.AllSettings }}{{ printf "%s = \"%v\"\n" $key $value }}{{end}}
`
