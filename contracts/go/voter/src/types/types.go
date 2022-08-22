package types

import (
	"errors"

	"github.com/CosmWasm/cosmwasm-go/std"
	stdTypes "github.com/CosmWasm/cosmwasm-go/std/types"

	"github.com/archway-network/voter/src/pkg"
	"github.com/archway-network/voter/src/state"
)

// Params is the contract parameters (state.Params representation).
type Params struct {
	// OwnerAddr is a contract owner canonical address (he can redeem funds).
	OwnerAddr string
	// NewVotingCost is a cost one should pay for creating a new voting.
	NewVotingCost string
	// VoteCost is a cost of a single vote.
	VoteCost string
	// IBCSendTimeout is a timeout for IBC send [ns].
	IBCSendTimeout uint64
}

// NewParamsFromState converts state.Params to Params.
func NewParamsFromState(api std.Api, params state.Params) (Params, error) {
	ownerAddr, err := api.HumanAddress(params.OwnerAddr)
	if err != nil {
		return Params{}, errors.New("ownerAddr: human convert: " + err.Error())
	}

	return Params{
		OwnerAddr:      ownerAddr,
		NewVotingCost:  params.NewVotingCost.String(),
		VoteCost:       params.VoteCost.String(),
		IBCSendTimeout: params.IBCSendTimeout,
	}, nil
}

// ValidateAndConvert performs object fields validation and converts to state.Params.
func (p Params) ValidateAndConvert(api std.Api, info stdTypes.MessageInfo) (state.Params, error) {
	if err := api.ValidateAddress(p.OwnerAddr); err != nil {
		return state.Params{}, errors.New("ownerAddr: " + err.Error())
	}

	ownerAddr, err := api.CanonicalAddress(p.OwnerAddr)
	if err != nil {
		return state.Params{}, errors.New("ownerAddr: canonical convert: " + err.Error())
	}

	if p.OwnerAddr != info.Sender {
		return state.Params{}, errors.New("ownerAddr: must EQ to senderAddr")
	}

	newVotingCostCoin, err := pkg.ParseCoinFromString(p.NewVotingCost)
	if err != nil {
		return state.Params{}, errors.New("newVotingCost: parsing coin: " + err.Error())
	}

	voteCost, err := pkg.ParseCoinFromString(p.VoteCost)
	if err != nil {
		return state.Params{}, errors.New("voteCost: parsing coin: " + err.Error())
	}

	if p.IBCSendTimeout == 0 {
		return state.Params{}, errors.New("ibcSendTimeout: must be GT 0")
	}

	return state.Params{
		OwnerAddr:      ownerAddr,
		NewVotingCost:  newVotingCostCoin,
		VoteCost:       voteCost,
		IBCSendTimeout: p.IBCSendTimeout,
	}, nil
}
