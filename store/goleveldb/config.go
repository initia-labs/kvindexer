package goleveldb

import (
	"errors"

	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

const (
	keyBloomFileterBits       = "bloom-filter-bits"
	keyBlockCacheCapacity     = "block-cache-capacity"
	keyOpenFilesCacheCapacity = "open-files-cache-capacity"
)

func DefaultConfig() *viper.Viper {
	vpr := viper.New()
	vpr.SetDefault(keyBloomFileterBits, 10) // same as the constant in cosmos-db
	vpr.SetDefault(keyBlockCacheCapacity, opt.DefaultBlockCacheCapacity)
	vpr.SetDefault(keyOpenFilesCacheCapacity, opt.DefaultOpenFilesCacheCapacity)
	return vpr
}

func ValidateConfig(vpr *viper.Viper) error {
	if vpr.GetInt(keyBloomFileterBits) < 0 {
		return errors.New("invalid bloom filter bits")
	}
	if vpr.GetInt(keyBlockCacheCapacity) < 0 {
		return errors.New("invalid block cache capacity")
	}
	if vpr.GetInt(keyOpenFilesCacheCapacity) < 0 {
		return errors.New("invalid open files cache capacity")
	}
	return nil
}

func ConvertOptions(vpr *viper.Viper) *opt.Options {
	return &opt.Options{
		Filter:                 filter.NewBloomFilter(viper.GetInt(keyBloomFileterBits)),
		BlockCacheCapacity:     vpr.GetInt(keyBlockCacheCapacity),
		OpenFilesCacheCapacity: vpr.GetInt(keyOpenFilesCacheCapacity),
	}
}
