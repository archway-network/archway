package gov

import (
	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

type Keeper struct {
	govKeeper govkeeper.Keeper
}

// NewKeeper creates a new Keeper instance.
func NewKeeper(govkeeper govkeeper.Keeper) Keeper {
	return Keeper{
		govKeeper: govkeeper,
	}
}

// GetVote returns the vote of a voter on a proposal.
func (k Keeper) GetVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) (govtypes.Vote, bool) {
	vote, err := k.govKeeper.Votes.Get(ctx, collections.Join(proposalID, voterAddr))
	if err != nil {
		return govtypes.Vote{}, false
	}
	return vote, true
}
