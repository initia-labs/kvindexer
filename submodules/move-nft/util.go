package move_nft

import (
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/initia-labs/kvindexer/submodules/move-nft/types"
)

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
