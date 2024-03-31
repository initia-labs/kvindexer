package nft

import (
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/initia-labs/kvindexer/submodule/nft/types"
)

// filter out events that are not related to move and have attribute named type_tag and data
//
//nolint:unused
func filterEvents(eventType string, events []abci.Event) (filtered []abci.Event) {

	for _, event := range events {
		if event.Type != eventType {
			continue
		}
		filtered = append(filtered, event)
	}

	return
}

func parseEvent(event abci.Event) types.EventWithAttributeMap {
	eventWithMap := types.EventWithAttributeMap{Event: &event, AttributesMap: make(map[string]string)}
	for _, attribute := range event.Attributes {
		eventWithMap.AttributesMap[attribute.GetKey()] = attribute.GetValue()
	}
	return eventWithMap
}

func filterAndParseEvent(eventType string, events []abci.Event) (filtered []types.EventWithAttributeMap) {
	for _, event := range events {
		if event.Type != eventType {
			continue
		}
		filtered = append(filtered, parseEvent(event))
	}
	return
}
