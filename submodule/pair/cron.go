package pair

import (
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/gofiber/fiber/v2"
	"github.com/initia-labs/kvindexer/config"
	"github.com/initia-labs/kvindexer/module/keeper"
	"github.com/spf13/cast"
)

var Cronjob = keeper.Cronjob{
	// wrapped with __ to avoid conflict with other cronjobs and mark it as a submodule's cronjob
	Tag:        fmt.Sprintf("__%s__", submoduleName),
	Initialize: pairCollectorInitializer,
	Job:        pairCollectorRunner,
}

var croncfg *cronConfig

const (
	bridgeIdConfigKey = "bridge_id"
	l1ChainId         = "l1_chain_id"
	l1LcdUrlConfigKey = "l1_lcd_url"
	l1QueryPatternKey = "l1_query_pattern"
)

type cronConfig struct {
	bridgeId       uint64
	l1ChainId      string
	ibcChannels    atomic.Value
	ibcNftChannels atomic.Value
	l1LcdUrl       string
	l1QueryPattern string
}

func getCronConfigFromSubmoduleConfig(smcfg config.SubmoduleConfig) (*cronConfig, error) {
	cfg := cronConfig{}

	// bridgeId base is 1, so if it's 0, it's not set
	cfg.bridgeId = cast.ToUint64(smcfg[bridgeIdConfigKey])
	if cfg.bridgeId == 0 {
		return nil, errors.New("bridge_id is required")
	}

	cfg.l1ChainId = cast.ToString(smcfg[l1ChainId])
	if cfg.l1ChainId == "" {
		return nil, errors.New("l1_chain_id is required")
	}

	cfg.l1QueryPattern = cast.ToString(smcfg[l1QueryPatternKey])
	if cfg.l1QueryPattern == "" {
		return nil, errors.New("l1_query_pattern is required")
	}

	cfg.l1LcdUrl = cast.ToString(smcfg[l1LcdUrlConfigKey])
	if cfg.l1LcdUrl == "" {
		return nil, errors.New("l1_lcd_url is required")
	}

	return &cfg, nil
}

func pairCollectorInitializer(keeper *keeper.Keeper, config config.CronjobConfig) error {
	// nop

	return nil
}

func pairCollectorRunner(keeper *keeper.Keeper, config config.CronjobConfig) error {
	client := fiber.AcquireClient()
	defer fiber.ReleaseClient(client)

	_ = collectOpTokenPairsFromL1(client, croncfg)
	_ = collectNftTokenPairsFromL1(client, croncfg)

	// return nil: it's cron
	return nil
}
