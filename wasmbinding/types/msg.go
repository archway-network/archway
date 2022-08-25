package types

import (
	"fmt"

	rewardsTypes "github.com/archway-network/archway/wasmbinding/rewards/types"
)

// Msg is a container for custom WASM messages (one of).
type Msg struct {
	// Rewards defines the x/rewards module specific sub-message.
	Rewards *rewardsTypes.Msg `json:"rewards,omitempty"`
}

// Validate validates the msg fields.
func (m Msg) Validate() error {
	cnt := 0

	if m.Rewards != nil {
		cnt++
	}

	if cnt != 1 {
		return fmt.Errorf("one and only one sub-message must be set")
	}

	return nil
}
