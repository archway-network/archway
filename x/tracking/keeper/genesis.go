package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/tracking/types"
)

// ExportGenesis exports the module genesis for the current block.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	txInfoLastID, txInfos := k.state.TxInfoState(ctx).Export()
	opInfoLastID, opInfos := k.state.ContractOpInfoState(ctx).Export()

	return types.NewGenesisState(
		txInfoLastID,
		txInfos,
		opInfoLastID,
		opInfos,
	)
}

// InitGenesis initializes the module genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, state *types.GenesisState) {
	k.state.TxInfoState(ctx).Import(state.TxInfoLastId, state.TxInfos)
	k.state.ContractOpInfoState(ctx).Import(state.ContractOpInfoLastId, state.ContractOpInfos)
}
