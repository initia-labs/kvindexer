package util

import (
	"errors"
	"slices"

	abci "github.com/cometbft/cometbft/abci/types"
)

func GetAttributeValue(event abci.Event, key string) (string, error) {
	i := slices.IndexFunc(event.Attributes, func(attr abci.EventAttribute) bool {
		return attr.Key == key
	})
	if i < 0 {
		return "", errors.New("not found")
	}
	return event.Attributes[i].Value, nil
}

func FilterEvent(eventType string, events []abci.Event) (filtered []abci.Event) {
	for _, event := range events {
		if event.Type != eventType {
			continue
		}
		filtered = append(filtered, event)
	}
	return
}
