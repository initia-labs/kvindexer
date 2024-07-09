package config

import (
	"github.com/spf13/viper"
)

type IndexerConfig struct {
	Enable        bool         `mapstructure:"indexer.enable"`
	CacheSize     int          `mapstructure:"indexer.cache-size"`
	BackendConfig *viper.Viper `mapstructure:"indexer.backend"`
}

const DefaultConfigTemplate = `
###############################################################################
###                              Indexer                                   ###
###############################################################################

[indexer]

# Enable defines whether the indexer is enabled.
enable = {{ .IndexerConfig.Enable }}

# CacheSize defines the size of the cache. (unit: bytes)
cache-size = {{ .IndexerConfig.CacheSize }}

# Backend defines the type of the backend store and its options.
# It should have a key-value pair named 'type', and the value should exist in store supported by cosmos-db.
# supported type: "goleveldb" only in current
[indexer.backend]
{{ range $key, $value := .IndexerConfig.BackendConfig.AllSettings }}{{ printf "%s = \"%v\"\n" $key $value }}{{end}}
`
