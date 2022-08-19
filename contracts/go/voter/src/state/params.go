package state

import (
	"errors"

	"github.com/CosmWasm/cosmwasm-go/std"
	stdTypes "github.com/CosmWasm/cosmwasm-go/std/types"
)

// ParamsKey is the storage key for Params state.
var ParamsKey = []byte("Params")

// Params defines the contract parameters (set via Instantiate endpoint).
type Params struct {
	// OwnerAddr is a contract owner canonical address (he can redeem funds).
	OwnerAddr stdTypes.CanonicalAddress
	// NewVotingCost is a cost one should pay for creating a new voting.
	NewVotingCost stdTypes.Coin
	// VoteCost is a cost of a single vote.
	VoteCost stdTypes.Coin
	// IBCSendTimeout is a timeout for IBC send [ns].
	IBCSendTimeout uint64
}

// GetParams returns Params state.
func GetParams(storage std.Storage) (params Params, retErr error) {
	defer func() {
		if retErr != nil {
			retErr = errors.New("params state get: " + retErr.Error())
		}
	}()

	bz := storage.Get(ParamsKey)
	if bz == nil {
		retErr = errors.New("not found")
		return
	}

	if err := params.UnmarshalJSON(bz); err != nil {
		retErr = errors.New("object JSON unmarshal")
		return
	}

	return
}

// SetParams sets Params state.
func SetParams(storage std.Storage, params Params) (retErr error) {
	defer func() {
		if retErr != nil {
			retErr = errors.New("params state set: " + retErr.Error())
		}
	}()

	bz, err := params.MarshalJSON()
	if err != nil {
		retErr = errors.New("object JSON marshal: " + err.Error())
		return
	}

	storage.Set(ParamsKey, bz)

	return
}
