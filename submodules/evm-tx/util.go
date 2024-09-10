package tx

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	initBech32RegexStr = "^init1(?:[a-z0-9]{38}|[a-z0-9]{58})$"
	initHexRegexStr    = "0x(?:[a-f1-9][a-f0-9]*){1,64}"
)

var (
	initBech32Regex = regexp.MustCompile(initBech32RegexStr)
	initHexRegex    = regexp.MustCompile(initHexRegexStr)
)

func findAllBech32Address(str string) []string {
	return initBech32Regex.FindAllString(str, -1)
}

func findAllHexAddress(str string) []string {
	return initHexRegex.FindAllString(str, -1)
}

// use it because i want to make this submodule not depend on move vm/module
func accAddressFromString(addrStr string) (addr sdk.AccAddress, err error) {
	if strings.HasPrefix(addrStr, "0x") {
		addrStr = strings.TrimPrefix(addrStr, "0x")

		// add padding
		if len(addrStr) <= 40 {
			addrStr = strings.Repeat("0", 40-len(addrStr)) + addrStr
		} else if len(addrStr) <= 64 {
			addrStr = strings.Repeat("0", 64-len(addrStr)) + addrStr
		} else {
			return nil, fmt.Errorf("invalid address string: %s", addrStr)
		}

		if addr, err = hex.DecodeString(addrStr); err != nil {
			return
		}
	} else if addr, err = sdk.AccAddressFromBech32(addrStr); err != nil {
		return
	}

	return
}

func convertContractAddressToBech32(addr string) (string, error) {
	accAddr, err := sdk.AccAddressFromHexUnsafe(strings.ToLower(strings.TrimPrefix(strings.TrimPrefix(addr, "0x"), "000000000000000000000000")))
	if err != nil {
		return "", err
	}
	return accAddr.String(), nil
}
