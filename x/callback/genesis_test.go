package callback_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/callback"
	"github.com/archway-network/archway/x/callback/types"
)

func TestExportGenesis(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1, e2eTesting.WithCallbackParams(123))
	keeper := chain.GetApp().Keepers.CallbackKeeper
	contractAdminAcc := chain.GetAccount(1)

	// Upload and instantiate contract
	codeID := chain.UploadContract(contractAdminAcc, "../../contracts/callback-test/artifacts/callback_test.wasm", wasmdTypes.DefaultUploadAccess)
	initMsg := CallbackContractInstantiateMsg{Count: 100}
	contractAddr, _ := chain.InstantiateContract(contractAdminAcc, codeID, contractAdminAcc.Address.String(), "callback_test", nil, initMsg)

	feesToPay, err := getCallbackRegistrationFees(chain)
	require.NoError(t, err)

	msgs := []sdk.Msg{}
	msgs = append(msgs, &types.MsgRequestCallback{
		ContractAddress: contractAddr.String(),
		JobId:           DECREMENT_JOBID,
		CallbackHeight:  chain.GetContext().BlockHeight() + 2,
		Sender:          contractAdminAcc.Address.String(),
		Fees:            feesToPay,
	})
	msgs = append(msgs, &types.MsgRequestCallback{
		ContractAddress: contractAddr.String(),
		JobId:           INCREMENT_JOBID,
		CallbackHeight:  chain.GetContext().BlockHeight() + 2,
		Sender:          contractAdminAcc.Address.String(),
		Fees:            feesToPay,
	})
	msgs = append(msgs, &types.MsgRequestCallback{
		ContractAddress: contractAddr.String(),
		JobId:           DONOTHING_JOBID,
		CallbackHeight:  chain.GetContext().BlockHeight() + 2,
		Sender:          contractAdminAcc.Address.String(),
		Fees:            feesToPay,
	})
	msgs = append(msgs, &types.MsgRequestCallback{
		ContractAddress: contractAddr.String(),
		JobId:           DECREMENT_JOBID,
		CallbackHeight:  chain.GetContext().BlockHeight() + 3,
		Sender:          contractAdminAcc.Address.String(),
		Fees:            feesToPay,
	})
	_, _, _, err = chain.SendMsgs(contractAdminAcc, true, msgs)
	require.NoError(t, err)

	exportedState := callback.ExportGenesis(chain.GetContext(), keeper)

	require.Equal(t, 4, len(exportedState.Callbacks))
	require.Equal(t, uint64(123), exportedState.Params.CallbackGasLimit)
}

func TestInitGenesis(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1)
	ctx, keeper := chain.GetContext(), chain.GetApp().Keepers.CallbackKeeper
	contractAddr := e2eTesting.GenContractAddresses(1)[0]
	validCoin := sdk.NewInt64Coin("stake", 10)

	genParams := types.Params{
		CallbackGasLimit:               1000000,
		MaxBlockReservationLimit:       1,
		MaxFutureReservationLimit:      1,
		BlockReservationFeeMultiplier:  math.LegacyZeroDec(),
		FutureReservationFeeMultiplier: math.LegacyZeroDec(),
	}
	err := keeper.SetParams(ctx, genParams)
	require.NoError(t, err)

	genstate := types.GenesisState{
		Params: genParams,
		Callbacks: []*types.Callback{
			{
				ContractAddress: contractAddr.String(),
				JobId:           1,
				CallbackHeight:  100,
				ReservedBy:      contractAddr.String(),
				FeeSplit: &types.CallbackFeesFeeSplit{
					TransactionFees:       &validCoin,
					BlockReservationFees:  &validCoin,
					FutureReservationFees: &validCoin,
					SurplusFees:           &validCoin,
				},
			},
		},
	}

	callback.InitGenesis(ctx, keeper, genstate)

	callbacks, err := keeper.GetAllCallbacks(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, len(callbacks)) // Ensuring callbacks are not imported

	params, err := keeper.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, genParams.CallbackGasLimit, params.CallbackGasLimit)
	require.Equal(t, genParams.MaxBlockReservationLimit, params.MaxBlockReservationLimit)
	require.Equal(t, genParams.MaxFutureReservationLimit, params.MaxFutureReservationLimit)
	require.Equal(t, genParams.BlockReservationFeeMultiplier, params.BlockReservationFeeMultiplier)
	require.Equal(t, genParams.FutureReservationFeeMultiplier, params.FutureReservationFeeMultiplier)
}
