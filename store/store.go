package store

import (
	"fmt"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/initia-labs/kvindexer/store/goleveldb"
	"github.com/spf13/viper"
)

// OpenDB returns an opened db based on the given configuration
func OpenDB(homeDir, name string, config *viper.Viper) (dbm.DB, error) {
	typ := dbm.BackendType(config.GetString(KeyType))
	switch typ {
	case dbm.GoLevelDBBackend:
		return goleveldb.NewDB(homeDir, name, config)
	default:
		return nil, fmt.Errorf("not supported backend type: %s", string(typ))
	}
}
