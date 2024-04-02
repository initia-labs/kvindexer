package dashboard

import (
	"errors"
	"github.com/initia-labs/kvindexer/config"
	"github.com/spf13/cast"
)

const (
	opBridgeIdKey = "op-bridge-id"
	l1DenomKey    = "l1-denom"
)

func checkConfig(conf config.SubmoduleConfig) error {
	opBridgeId := cast.ToUint64(conf[opBridgeIdKey])
	if opBridgeId == 0 {
		return errors.New("op-bridge-id is required")
	}
	l1Denom := cast.ToString(conf[l1DenomKey])
	if l1Denom == "" {
		return errors.New("l1-denom is required")
	}
	return nil
}
