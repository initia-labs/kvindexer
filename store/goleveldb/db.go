package goleveldb

import (
	dbm "github.com/cosmos/cosmos-db"
	"github.com/spf13/viper"
)

func NewDB(homeDir, name string, config *viper.Viper) (*dbm.GoLevelDB, error) {
	return dbm.NewGoLevelDBWithOpts(name, homeDir, ConvertOptions(config))
}
