package store

import (
	"fmt"

	dbm "github.com/cosmos/cosmos-db"

	"github.com/initia-labs/kvindexer/store/goleveldb"
	"github.com/spf13/viper"
)

const (
	KeyType = "type"
)

func DefaultConfig() *viper.Viper {
	// use goleveldb as default
	vpr := goleveldb.DefaultConfig()
	vpr.SetDefault(KeyType, dbm.GoLevelDBBackend)
	return vpr
}

func ValidateConfig(vpr *viper.Viper) error {
	typ := dbm.BackendType(vpr.GetString(KeyType))
	switch typ {
	case dbm.GoLevelDBBackend:
		return goleveldb.ValidateConfig(vpr)
	default:
		return fmt.Errorf("not supported backend type: %s", vpr.GetString("type"))
	}
}
