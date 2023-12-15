package callback_test

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	e2eTesting "github.com/archway-network/archway/e2e/testing"
	callbackKeeper "github.com/archway-network/archway/x/callback/keeper"
	"github.com/archway-network/archway/x/callback/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

func TestEndBlocker(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1)
	ctx, keeper := chain.GetContext().WithBlockHeight(1000), chain.GetApp().Keepers.CallbackKeeper
	msgServer := callbackKeeper.NewMsgServer(keeper)
	contractAdminAcc := chain.GetAccount(0)

	chain.GetApp().Keepers.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("stake", 3500000000)))
	chain.GetApp().Keepers.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, contractAdminAcc.Address, sdk.NewCoins(sdk.NewInt64Coin("stake", 3500000000)))

	// Upload and instantiate contract
	codeID := chain.UploadContract(contractAdminAcc, "../../contracts/callback-test/artifacts/callback_test.wasm", wasmdTypes.DefaultUploadAccess)
	initMsg := CallbackContractInstantiateMsg{Count: 100}
	contractAddr, _ := chain.InstantiateContract(contractAdminAcc, codeID, contractAdminAcc.Address.String(), "callback_test", nil, initMsg)

	chain.NextBlock(1 * time.Second)

	reqMsg := &types.MsgRequestCallback{
		ContractAddress: contractAddr.String(),
		JobId:           1,
		CallbackHeight:  1030,
		Sender:          contractAdminAcc.Address.String(),
		Fees:            sdk.NewInt64Coin("stake", 3500000000),
	}
	_, err := msgServer.RequestCallback(sdk.WrapSDKContext(ctx), reqMsg)
	require.NoError(t, err)
	_, err = keeper.GetCallback(ctx, reqMsg.CallbackHeight, reqMsg.ContractAddress, reqMsg.JobId)
	require.NoError(t, err)
}

type CallbackContractInstantiateMsg struct {
	Count int32 `json:"count"`
}

func (msg CallbackContractInstantiateMsg) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Count string `json:"count"`
	}{
		Count: strconv.Itoa(int(msg.Count)),
	})
}
