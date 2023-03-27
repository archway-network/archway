package types

import (
	"errors"
	"strconv"

	stdTypes "github.com/CosmWasm/cosmwasm-go/std/types"

	"github.com/archway-network/voter/src/pkg"
	archwayCustomTypes "github.com/archway-network/voter/src/pkg/archway/custom"
)

// MsgExecute is handled by the Execute entrypoint.
type MsgExecute struct {
	// Release sends raised by the contract funds to its creator.
	Release *struct{} `json:",omitempty"`
	// NewVoting creates a new voting.
	NewVoting *NewVotingRequest `json:",omitempty"`
	// Vote append a new vote to an existing voting.
	Vote *VoteRequest `json:",omitempty"`
	// SendIBCVote append a new vote to an existing voting over IBC.
	SendIBCVote *SendIBCVoteRequest `json:",omitempty"`

	// CustomCustom calls WASM bindings with a custom msg.
	CustomCustom stdTypes.RawMessage `json:",omitempty"`
	// CustomUpdateMetadata calls WASM bindings UpdateMetadata custom msg.
	CustomUpdateMetadata *archwayCustomTypes.UpdateContractMetadataRequest `json:",omitempty"`
	// CustomWithdrawRewards calls WASM bindings WithdrawRewards custom msg.
	CustomWithdrawRewards *archwayCustomTypes.WithdrawRewardsRequest `json:",omitempty"`
	// Fail causes the contract to fail
	Fail *struct{} `json:",omitempty"`
	// ReplyOnError causes the contract to send a message to the provided destination.
	ReplyOnError *string `json:",omitempty"`
}

// ReleaseResponse defines MsgExecute.Release response.
type ReleaseResponse struct {
	ReleasedAmount []stdTypes.Coin
}

type (
	// NewVotingRequest defines MsgExecute.NewVoting request.
	NewVotingRequest struct {
		// Name is the new voting name.
		Name string
		// VoteOptions are voting options.
		VoteOptions []string
		// Duration is the voting duration [ns].
		Duration uint64
	}

	// NewVotingResponse defines MsgExecute.NewVoting response.
	NewVotingResponse struct {
		// VotingID is a unique voting ID.
		VotingID uint64
	}
)

// Validate performs object fields validation.
func (r NewVotingRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name: empty")
	}

	if len(r.VoteOptions) == 0 {
		return errors.New("voteOptions: empty")
	}
	for i, opt := range r.VoteOptions {
		if opt == "" {
			return errors.New("voteOptions [" + strconv.Itoa(i) + "]: empty")
		}
	}

	if r.Duration == 0 {
		return errors.New("duration: must be GT 0")
	}

	return nil
}

// VoteRequest defines MsgExecute.Vote request.
type VoteRequest struct {
	// ID is a unique voting ID.
	ID uint64
	// Option is a voting option.
	Option string
	// Vote is a voting option (yes / no).
	Vote string
}

// Validate performs object fields validation.
func (r VoteRequest) Validate() error {
	if r.Option == "" {
		return errors.New("option: empty")
	}

	switch r.Vote {
	case "yes":
	case "no":
	default:
		return errors.New("unknown vote enum value (yes/no is expected)")
	}

	return nil
}

// SendIBCVoteRequest defines MsgExecute.SendIBCVoteRequest request.
type SendIBCVoteRequest struct {
	VoteRequest

	// ChannelID is an IBC destination channel.
	ChannelID string
}

// Validate performs object fields validation.
func (r SendIBCVoteRequest) Validate() error {
	if err := r.VoteRequest.Validate(); err != nil {
		return err
	}

	if err := pkg.ValidateChannelID(r.ChannelID); err != nil {
		return errors.New("channelID: " + err.Error())
	}

	return nil
}
