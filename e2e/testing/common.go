package e2eTesting

import (
	abci "github.com/tendermint/tendermint/abci/types"
)

// GetStringEventAttribute returns TX response event attribute string value by type and attribute key.
func GetStringEventAttribute(events []abci.Event, eventType, attrKey string) string {
	for _, event := range events {
		if event.Type != eventType {
			continue
		}

		for _, attr := range event.Attributes {
			if string(attr.Key) != attrKey {
				continue
			}

			return string(attr.Value)
		}
	}

	return ""
}
