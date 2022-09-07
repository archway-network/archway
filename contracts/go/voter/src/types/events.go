package types

import (
	"strconv"

	stdTypes "github.com/CosmWasm/cosmwasm-go/std/types"
)

const (
	EventTypeRelease         = "release"
	EventTypeNewVoting       = "new_voting"
	EventTypeVote            = "vote"
	EventTypeIBCVoteSent     = "ibc_vote_sent"
	EventTypeIBCVoteReceived = "ibc_vote_received"

	EventTypeNewVotingCostChanged = "new_voting_cost_change"
	EventTypeVoteCostChanged      = "vote_cost_change"

	EventAttrKeySender       = "sender"
	EventAttrKeyVotingID     = "voting_id"
	EventAttrKeyVoteOption   = "vote_option"
	EventAttrKeyVoteDecision = "vote_decision"
	EventAttrKeyOldCost      = "old_cost"
	EventAttrKeyNewCost      = "new_cost"
)

// NewEventRelease creates a new Event on funds release.
func NewEventRelease(ownerAddr string) stdTypes.Event {
	return stdTypes.Event{
		Type: EventTypeRelease,
		Attributes: []stdTypes.EventAttribute{
			{
				Key:   EventAttrKeySender,
				Value: ownerAddr,
			},
		},
	}
}

// NewEventVotingCreated creates a new Event on voting creation.
func NewEventVotingCreated(creatorAddr string, votingID uint64) stdTypes.Event {
	return stdTypes.Event{
		Type: EventTypeNewVoting,
		Attributes: []stdTypes.EventAttribute{
			{
				Key:   EventAttrKeySender,
				Value: creatorAddr,
			},
			{
				Key:   EventAttrKeyVotingID,
				Value: strconv.FormatUint(votingID, 10),
			},
		},
	}
}

// NewEventVoteAdded creates a new Event on vote event.
func NewEventVoteAdded(senderAddr string, votingID uint64, option, vote string) stdTypes.Event {
	return stdTypes.Event{
		Type: EventTypeVote,
		Attributes: []stdTypes.EventAttribute{
			{
				Key:   EventAttrKeySender,
				Value: senderAddr,
			},
			{
				Key:   EventAttrKeyVotingID,
				Value: strconv.FormatUint(votingID, 10),
			},
			{
				Key:   EventAttrKeyVoteOption,
				Value: option,
			},
			{
				Key:   EventAttrKeyVoteDecision,
				Value: vote,
			},
		},
	}
}

// NewEventIBCVoteSent creates a new Event on IBC vote send event.
func NewEventIBCVoteSent(fromAddr string, votingID uint64, option, vote string) stdTypes.Event {
	event := NewEventVoteAdded(fromAddr, votingID, option, vote)
	event.Type = EventTypeIBCVoteSent

	return event
}

// NewEventIBCVoteAdded creates a new Event on IBC vote event.
func NewEventIBCVoteAdded(fromAddr string, votingID uint64, option, vote string) stdTypes.Event {
	event := NewEventVoteAdded(fromAddr, votingID, option, vote)
	event.Type = EventTypeIBCVoteReceived

	return event
}

// NewEventNewVotingCostChanged creates a new Event on new voting cost change sudo event.
func NewEventNewVotingCostChanged(oldCost, newCost stdTypes.Coin) stdTypes.Event {
	return stdTypes.Event{
		Type: EventTypeNewVotingCostChanged,
		Attributes: []stdTypes.EventAttribute{
			{
				Key:   EventAttrKeyOldCost,
				Value: oldCost.String(),
			},
			{
				Key:   EventAttrKeyNewCost,
				Value: newCost.String(),
			},
		},
	}
}

// NewEventVoteCostChanged creates a new Event on vote cost change sudo event.
func NewEventVoteCostChanged(oldCost, newCost stdTypes.Coin) stdTypes.Event {
	return stdTypes.Event{
		Type: EventTypeVoteCostChanged,
		Attributes: []stdTypes.EventAttribute{
			{
				Key:   EventAttrKeyOldCost,
				Value: oldCost.String(),
			},
			{
				Key:   EventAttrKeyNewCost,
				Value: newCost.String(),
			},
		},
	}
}
