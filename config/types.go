package config

import (
	"github.com/spf13/viper"
)

type SubmoduleConfig map[string]interface{}
type CronjobConfig map[string]interface{}

type IndexerConfig struct {
	Enable            bool                       `mapstructure:"indexer.enable"`
	BackendConfig     *viper.Viper               `mapstructure:"indexer.backend"`
	EnabledSubmodules []string                   `mapstructure:"indexer.enabled-submodules"`
	EnabledCronJobs   []string                   `mapstructure:"indexer.enabled-cronjobs"`
	SubmoduleConfigs  map[string]SubmoduleConfig // key: submodule name, value: kv pair as a map[string]interface{}
	CronjobConfigs    map[string]CronjobConfig   // key: cronjob name, value: kv pair as a map[string]interface{}
	CacheSize         uint                       `mapstructure:"indexer.cache-size"`
}

const DefaultConfigTemplate = `
###############################################################################
###                              Indexer                                   ###
###############################################################################

[indexer]

# Enable defines whether the indexer is enabled.
enable = {{ .IndexerConfig.Enable }}

# cache size defines how many objects shoud be stored.
cache-size = 1000000


# Enable defines a list of the indexer submodules should be enabled.
enabled-submodules = "{{ range .IndexerConfig.EnabledSubmodules}}{{ printf "%q, " . }}{{end}}"

# Enable defines a list of the cronjobs should be enabled.
enabled-cronjobs = "{{ range .IndexerConfig.EnabledCronJobs }}{{ printf "%q, " . }}{{end}}"

# Backend defines the type of the backend store and its options.
# It should have a key-value pair named 'type', and the value should exist in store supported by cosmos-db.
# supported type: "goleveldb" only in current
[indexer.backend]
{{ range $key, $value := .IndexerConfig.BackendConfig.AllSettings }}{{ printf "%s = \"%v\"\n" $key $value }}{{end}}

# Submodules array defines the configuration for each submodule.
# It should have a key-value pair named 'name', and the value should exist in enabled-submodules.
# If the value of 'name' is not exist in enabled-submodules, it will be ignored.
[[indexer.submodules]]
name = "nop"
key1 = true
key2 = 83
key3 = "string"

# Cronjobs array defines the configuration for each cronjob.
# It should have a key-value pair named 'name', and the value should exist in enabled-cronjobs.
# Also it should have a key-value pair named 'pattern' for cronjob pattern.
# If the value of 'name' is not exist in enabled-cronjobs, it will be ignored.
[[indexer.cronjobs]]
name = "cronjob1"
pattern = "0 0 1 * *"
`
