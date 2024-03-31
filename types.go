package indexer

import (
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/initia-labs/indexer/v2/config"
	"github.com/initia-labs/indexer/v2/module/keeper"
)

type Indexer struct {
	config *config.IndexerConfig
	keeper *keeper.Keeper
	logger log.Logger
}

type IndexableApplication interface {
	GetIndexerKeeper() *keeper.Keeper
	GetBaseApp() *baseapp.BaseApp
	GetKeys() map[string]*storetypes.KVStoreKey
}
