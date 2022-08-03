package integration

import (
	"encoding/json"
	"testing"

	voterTypes "github.com/CosmWasm/cosmwasm-go/example/voter/src/types"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/gastracker"
	"github.com/archway-network/archway/x/gastracker/common"
)

func TestRewardsCollection(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1)
	// TODO: this test can be done better but for the sake of simplicity lets keep it like this for now
	const blocks int64 = 2
	var inflation = sdk.NewInt64Coin("stake", 103)

	params, err := gastracker.NewQueryClient(chain.Client()).Params(chain.GetContext().Context(), &gastracker.QueryParamsRequest{})
	require.NoError(t, err)

	totalInflation := sdk.NewCoin(
		inflation.Denom, (inflation.Amount.ToDec().Mul(params.Params.DappInflationRewardsRatio)).MulInt64(blocks).TruncateInt().SubRaw(1)) // we're subbing a meaningless residual due to loss of precision

	gasTrackerBalance := chain.GetBalance(authtypes.NewModuleAddress(gastracker.ModuleName))
	require.Equal(t,
		gasTrackerBalance.String(),
		sdk.NewCoins(totalInflation).String(),
	)

	contractAddr := uploadAndInstantiateContract(chain)

	msg := &voterTypes.MsgExecute{
		NewVoting: &voterTypes.NewVotingRequest{
			Name:        "hello",
			VoteOptions: []string{"idk"},
			Duration:    100,
		},
	}
	txFees := sdk.NewCoins(sdk.NewInt64Coin("stake", 1000))
	chain.SendMsgs(chain.GetAccount(0), true, []sdk.Msg{&wasmtypes.MsgExecuteContract{
		Sender:   chain.GetAccount(0).Address.String(),
		Contract: contractAddr.String(),
		Msg:      jsonMarshal(t, msg),
		Funds:    sdk.NewCoins(sdk.NewInt64Coin("stake", 100)),
	}},
		e2eTesting.WithMsgFees(txFees...),
	)

	balance := chain.GetBalance(authtypes.NewModuleAddress(gastracker.ModuleName))
	totalInflation = sdk.NewCoin(
		inflation.Denom, (inflation.Amount.ToDec().Mul(params.Params.DappInflationRewardsRatio)).MulInt64(5).TruncateInt().SubRaw(3)) // we're subbing a meaningless residual due to loss of precision

	_, dappRewardFees := common.SplitCoins(params.Params.DappTxFeeRebateRatio, txFees)
	require.Equal(t,
		dappRewardFees,
		balance.Sub(sdk.NewCoins(totalInflation)), // remove inflation
	)
}

func uploadAndInstantiateContract(chain *e2eTesting.TestChain) sdk.AccAddress {
	owner := chain.GetAccount(0)
	id := chain.UploadContract(owner, "../../../e2e/contracts/voter.wasm", wasmtypes.DefaultUploadAccess)
	addr, _ := chain.InstantiateContract(owner, id, "", "voter", nil, voterTypes.MsgInstantiate{Params: voterTypes.Params{
		OwnerAddr:      owner.Address.String(),
		NewVotingCost:  "100stake",
		VoteCost:       "100stake",
		IBCSendTimeout: 10_000_000,
	}})

	return addr
}

func jsonMarshal(t *testing.T, msg interface{}) []byte {
	b, err := json.Marshal(msg)
	require.NoError(t, err)
	return b
}
