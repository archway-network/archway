package types

import (
	"fmt"

	rewardsTypes "github.com/archway-network/archway/wasmbinding/rewards/types"
)

// Query is a container for custom WASM queries (one of).
type Query struct {
	// Rewards defines the x/rewards module specific sub-query.
	Rewards *rewardsTypes.Query `json:"rewards,omitempty"`
}

// Validate validates the query fields.
func (q Query) Validate() error {
	cnt := 0

	if q.Rewards != nil {
		cnt++
	}

	if cnt != 1 {
		return fmt.Errorf("one and only one sub-query must be set")
	}

	return nil
}
