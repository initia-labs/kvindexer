package config

import (
	"fmt"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/spf13/cast"
	"github.com/spf13/viper"

	"github.com/initia-labs/kvindexer/store"
)

const (
	flagIndexerEnable        = "indexer.enable"
	flagIndexerBackend       = "indexer.backend"
	flagIndexerCacheCapacity = "indexer.cache-capacity"
)

func NewConfig(appOpts servertypes.AppOptions) (*IndexerConfig, error) {
	cfg := &IndexerConfig{}

	cfg.Enable = cast.ToBool(appOpts.Get(flagIndexerEnable))
	if !cfg.Enable {
		return cfg, nil
	}
	cfg.CacheCapacity = cast.ToInt(appOpts.Get(flagIndexerCacheCapacity))

	cfg.BackendConfig = viper.New()
	err := cfg.BackendConfig.MergeConfigMap(cast.ToStringMap(appOpts.Get(flagIndexerBackend)))
	if err != nil {
		return nil, fmt.Errorf("failed to merge backend config: %w", err)
	}

	cfg.CacheCapacity = cast.ToInt(appOpts.Get(flagIndexerCacheCapacity))

	return cfg, nil
}

func (c IndexerConfig) Validate() error {
	if !c.Enable {
		return nil
	}

	if c.CacheCapacity == 0 {
		return fmt.Errorf("cache capacity must be greater than 0")
	}

	if c.BackendConfig == nil {
		return fmt.Errorf("backend config must be set")
	}

	return nil
}

func (c IndexerConfig) IsEnabled() bool {
	return c.Enable
}

func DefaultConfig() IndexerConfig {
	return IndexerConfig{
		Enable:        true,
		CacheCapacity: 500, // 500 MiB
		BackendConfig: store.DefaultConfig(),
	}
}
