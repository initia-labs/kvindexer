package config

import (
	"fmt"
	"slices"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/spf13/cast"
	"github.com/spf13/viper"

	"github.com/initia-labs/indexer/v2/store"
)

const (
	flagIndexerEnable            = "indexer.enable"
	flagIndexerBackend           = "indexer.backend"
	flagIndexerEnabledSubmodules = "indexer.enabled-submodules"
	flagIndexerEnabledCronjobs   = "indexer.enabled-cronjobs"
	flagIndexerSubmodules        = "indexer.submodules"
	flagIndexerCronjobs          = "indexer.cronjobs"
)

func NewConfig(appOpts servertypes.AppOptions) (*IndexerConfig, error) {
	cfg := &IndexerConfig{}

	cfg.Enable = cast.ToBool(appOpts.Get(flagIndexerEnable))
	if !cfg.Enable {
		return cfg, nil
	}

	cfg.EnabledSubmodules = cast.ToStringSlice(appOpts.Get(flagIndexerEnabledSubmodules))
	slices.Sort(cfg.EnabledSubmodules)

	cfg.SubmoduleConfigs = map[string]SubmoduleConfig{}
	svcCfgs := cast.ToSlice(appOpts.Get(flagIndexerSubmodules))
	for _, v := range svcCfgs {
		v := cast.ToStringMap(v)
		name, ok := v["name"]
		if !ok {
			return nil, fmt.Errorf("submodule name is not set")
		}
		cfg.SubmoduleConfigs[cast.ToString(name)] = cast.ToStringMap(v)
	}

	cfg.EnabledCronJobs = cast.ToStringSlice(appOpts.Get(flagIndexerEnabledCronjobs))
	slices.Sort(cfg.EnabledCronJobs)

	cfg.CronjobConfigs = map[string]CronjobConfig{}
	cjCfgs := cast.ToSlice(appOpts.Get(flagIndexerCronjobs))
	for _, v := range cjCfgs {
		v := cast.ToStringMap(v)
		name, ok := v["name"]
		if !ok {
			return nil, fmt.Errorf("cronjob name is not set")
		}
		cfg.CronjobConfigs[cast.ToString(name)] = cast.ToStringMap(v)
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

	// validate backend config
	if err := store.ValidateConfig(c.BackendConfig); err != nil {
		return fmt.Errorf("failed to validate backend config: %w", err)
	}

	if !slices.IsSorted(c.EnabledSubmodules) {
		return fmt.Errorf("enabled submodules must be sorted")
	}
	for _, svc := range c.EnabledSubmodules {
		if _, found := c.SubmoduleConfigs[svc]; !found {
			return fmt.Errorf("submodule %s is enabled but not configured", svc)
		}
	}

	if !slices.IsSorted(c.EnabledCronJobs) {
		return fmt.Errorf("enabled cronjobs must be sorted")
	}
	for _, cj := range c.EnabledCronJobs {
		cfg, found := c.CronjobConfigs[cj]
		if found {
			return fmt.Errorf("cronjob %s is enabled but not configured", cj)
		}
		pattern, ok := cfg["pattern"].(string)
		if !ok || pattern == "" {
			return fmt.Errorf("cronjob %s is enabled but pattern is not set", cj)
		}
	}

	return nil
}

func (c IndexerConfig) IsEnabled() bool {
	return c.Enable
}

func (c IndexerConfig) IsEnabledSubmodule(name string) bool {
	_, found := slices.BinarySearch(c.EnabledSubmodules, name)
	return found
}

func (c IndexerConfig) IsEnabledCronjob(tag string) bool {
	_, found := slices.BinarySearch(c.EnabledCronJobs, tag)
	return found
}

func DefaultConfig() IndexerConfig {
	return IndexerConfig{
		Enable:            false,
		BackendConfig:     store.DefaultConfig(),
		EnabledSubmodules: []string{},
		EnabledCronJobs:   []string{},
		SubmoduleConfigs:  map[string]SubmoduleConfig{},
		CronjobConfigs:    map[string]CronjobConfig{},
	}
}
