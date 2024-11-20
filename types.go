package indexer

import (
	"cosmossdk.io/log"
	"github.com/initia-labs/kvindexer/config"
	"github.com/initia-labs/kvindexer/x/kvindexer/keeper"
)

type Indexer struct {
	config *config.IndexerConfig
	keeper *keeper.Keeper
	logger log.Logger
}
