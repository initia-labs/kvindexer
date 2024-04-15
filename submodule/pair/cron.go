package pair

import (
	"fmt"
	"sync/atomic"

	"github.com/gofiber/fiber/v2"
	"github.com/initia-labs/kvindexer/module/keeper"
)

var Cronjob = keeper.Cronjob{
	// wrapped with __ to avoid conflict with other cronjobs and mark it as a submodule's cronjob
	Tag:        fmt.Sprintf("__%s__", submoduleName),
	Initialize: pairCollectorInitializer,
	Job:        pairCollectorRunner,
}

var croncfg *cronConfig

const (
	bridgeIdConfigKey = "op-bridge-id"
	l1ChainId         = "l1-chain-id"
	l1LcdUrlConfigKey = "l1-lcd-url"
	l1QueryPatternKey = "l1-query-pattern"
)

type cronConfig struct {
	bridgeId       uint64
	l1ChainId      string
	ibcChannels    atomic.Value
	ibcNftChannels atomic.Value
	l1LcdUrl       string
	l1QueryPattern string
}

func getCronConfigFromSubmoduleConfig() (*cronConfig, error) {
	cfg := cronConfig{}

	cfg.bridgeId = 0
	cfg.l1ChainId = ""
	cfg.l1LcdUrl = ""
	cfg.l1QueryPattern = ""

	return &cfg, nil
}

func pairCollectorInitializer(keeper *keeper.Keeper) error {
	// nop

	return nil
}

func pairCollectorRunner(keeper *keeper.Keeper) error {
	client := fiber.AcquireClient()
	defer fiber.ReleaseClient(client)

	_ = collectOpTokenPairsFromL1(client, croncfg)
	_ = collectNftTokenPairsFromL1(client, croncfg)

	// return nil: it's cron
	return nil
}
