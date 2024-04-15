package config

import (
	"fmt"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/spf13/cast"
	"github.com/spf13/viper"

	"github.com/initia-labs/kvindexer/store"
)

const (
	flagIndexerEnable    = "indexer.enable"
	flagIndexerBackend   = "indexer.backend"
	flagIndexerCacheSize = "indexer.cache-size"
)

func NewConfig(appOpts servertypes.AppOptions) (*IndexerConfig, error) {
	cfg := &IndexerConfig{}

	cfg.Enable = cast.ToBool(appOpts.Get(flagIndexerEnable))
	if !cfg.Enable {
		return cfg, nil
	}

	cfg.BackendConfig = viper.New()
	err := cfg.BackendConfig.MergeConfigMap(cast.ToStringMap(appOpts.Get(flagIndexerBackend)))
	if err != nil {
		return nil, fmt.Errorf("failed to merge backend config: %w", err)
	}

	return cfg, nil
}

func (c IndexerConfig) Validate() error {
	if !c.Enable {
		return nil
	}

	return nil
}

func (c IndexerConfig) IsEnabled() bool {
	return c.Enable
}

func DefaultConfig() IndexerConfig {
	return IndexerConfig{
		Enable:        false,
		CacheSize:     1_000_000,
		BackendConfig: store.DefaultConfig(),
	}
}
