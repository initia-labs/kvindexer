package evm_nft

import (
	"regexp"
	"sort"
	"strings"

	abci "github.com/cometbft/cometbft/abci/types"
)

// regexStripNonAlnum is used to strip non-alphanumeric characters from the collection name.
var regexStripNonAlnum = regexp.MustCompile("[^a-zA-Z0-9]+")

func filterEvents(events []abci.Event, eventType []string) (filtered []abci.Event) {
	eventTypeMap := make(map[string]bool)
	for _, eventType := range eventType {
		eventTypeMap[eventType] = true
	}

	for _, event := range events {
		if isTarget, found := eventTypeMap[event.Type]; found && isTarget {
			filtered = append(filtered, event)
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
