package simulation

import (
	"github.com/archway-network/archway/x/gastracker/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

func RandomizedGenState(simState *module.SimulationState) {
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&types.GenesisState{})
}
