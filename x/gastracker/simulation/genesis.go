package simulation

import (
	"github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/archway-network/archway/x/gastracker"
)

func RandomizedGenState(simState *module.SimulationState) {
	simState.GenState[gastracker.ModuleName] = simState.Cdc.MustMarshalJSON(&types.GenesisState{})
}
