package genesis

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/archway-network/archway/app"
	archwayappparams "github.com/archway-network/archway/app/params"
	"github.com/archway-network/archway/x/common/testutil/testapp"
	"github.com/archway-network/archway/x/oracle/denoms"
)

/*
	NewTestGenesisState returns 'NewGenesisState' using the default

genesis as input. The blockchain genesis state is represented as a map from module
identifier strings to raw json messages.
*/
func NewTestGenesisState(encodingConfig archwayappparams.EncodingConfig) app.GenesisState {
	codec := encodingConfig.Marshaler
	genState := app.NewDefaultGenesisState(codec)

	// Set short voting period to allow fast gov proposals in tests
	var govGenState govtypes.GenesisState
	codec.MustUnmarshalJSON(genState[gov.ModuleName], &govGenState)
	*govGenState.Params.VotingPeriod = time.Second * 20
	govGenState.Params.MinDeposit = sdk.NewCoins(sdk.NewInt64Coin(denoms.NIBI, 1_000_000)) // min deposit of 1 NIBI
	genState[gov.ModuleName] = codec.MustMarshalJSON(&govGenState)

	testapp.SetDefaultSudoGenesis(genState)

	return genState
}
