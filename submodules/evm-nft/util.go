package evm_nft

import (
	"sort"
	"strings"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/initia-labs/kvindexer/submodules/evm-nft/types"
)

func parseEvent(event abci.Event) types.EventWithAttributeMap {
	eventWithMap := types.EventWithAttributeMap{Event: &event, AttributesMap: make(map[string]string)}
	for _, attribute := range event.Attributes {
		eventWithMap.AttributesMap[attribute.GetKey()] = attribute.GetValue()
	}
	return eventWithMap
}

func filterAndParseEvent(events []abci.Event, eventTypes []string) (filtered []types.EventWithAttributeMap) {
	for _, event := range events {
		for _, eventType := range eventTypes {
			if event.Type != eventType {
				continue
			}
			filtered = append(filtered, parseEvent(event))
		}
	}
	return
}

// appendString appends two strings with a comma separator.
func appendString(s1, s2 string) string {
	strs := expandString([]string{s1, s2})

	strmap := make(map[string]bool)
	for _, str := range strs {
		strmap[str] = true
	}

	uniquestrs := make([]string, 0, len(strmap))
	for str := range strmap {
		if str == "" {
			continue
		}
		uniquestrs = append(uniquestrs, str)
	}
	sort.Strings(uniquestrs)
	return strings.Join(uniquestrs, ",")
}

func expandString(s []string) (res []string) {
	for _, v := range s {
		res = append(res, strings.Split(v, ",")...)
	}
	return res
}

func stripNonAlnum(in string) string {
	return regexStripNonAlnum.ReplaceAllString(in, "")
}
