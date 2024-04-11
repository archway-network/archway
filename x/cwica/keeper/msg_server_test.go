package keeper_test

import (
	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/cwica/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
)

// TestKeeper_RegisterInterchainAccount tests the RegisterInterchainAccount gRPC service method
func (s *KeeperTestSuite) TestRegisterInterchainAccount() {
	ctx, cwicaKeeper := s.chain.GetContext().WithBlockHeight(100), s.chain.GetApp().Keepers.CWICAKeeper
	wmKeeper, icaCtrlKeeper, connectionKeeper := testutils.NewMockContractViewer(), testutils.NewMockICAControllerKeeper(), testutils.NewMockConnectionlKeeper()
	cwicaKeeper.SetWasmKeeper(wmKeeper)
	cwicaKeeper.SetICAControllerKeeper(icaCtrlKeeper)
	cwicaKeeper.SetConnectionKeeper(connectionKeeper)
	contractAddress := e2eTesting.GenContractAddresses(1)[0]
	contractAdminAcc := s.chain.GetAccount(0)
	goCtx := sdk.WrapSDKContext(ctx)

	// TEST CASE 1: invalid contract address
	resp, err := cwicaKeeper.RegisterInterchainAccount(goCtx, &types.MsgRegisterInterchainAccount{})
	s.Require().ErrorContains(err, "failed to parse address")
	s.Require().Nil(resp)

	// TEST CASE 2: contract address is not a registered contract
	msgRegAcc := types.MsgRegisterInterchainAccount{
		ContractAddress: contractAddress.String(),
		ConnectionId:    "connection-0",
	}
	s.Require().False(wmKeeper.HasContractInfo(ctx, contractAddress))
	resp, err = cwicaKeeper.RegisterInterchainAccount(goCtx, &msgRegAcc)
	s.Require().ErrorContains(err, "is not a contract address")
	s.Require().Nil(resp)

	// TEST CASE 3: ibc connection not found for counterparty
	wmKeeper.AddContractAdmin(
		contractAddress.String(),
		contractAdminAcc.Address.String(),
	)
	s.Require().True(wmKeeper.HasContractInfo(ctx, contractAddress))
	resp, err = cwicaKeeper.RegisterInterchainAccount(goCtx, &msgRegAcc)
	s.Require().ErrorContains(err, "failed to get connection for counterparty")
	s.Require().Nil(resp)

	// TEST CASE 4: failed to register interchain account - e.g ica controller module disabled
	connectionKeeper.SetTestStateConnection()
	resp, err = cwicaKeeper.RegisterInterchainAccount(goCtx, &msgRegAcc)
	s.Require().ErrorContains(err, "failed to create RegisterInterchainAccount")
	s.Require().Nil(resp)

	// TEST CASE 5: successfully registered interchain account
	icaCtrlKeeper.SetTestStateRegisterInterchainAccount(false)
	resp, err = cwicaKeeper.RegisterInterchainAccount(goCtx, &msgRegAcc)
	s.Require().NoError(err)
	s.Require().Equal(types.MsgRegisterInterchainAccountResponse{}, *resp)
}

// TestKeeper_SendTx tests the SendTx gRPC service method
func (s *KeeperTestSuite) TestSendTx() {
	ctx, cwicaKeeper := s.chain.GetContext().WithBlockHeight(100), s.chain.GetApp().Keepers.CWICAKeeper
	wmKeeper, icaCtrlKeeper, channelKeeper := testutils.NewMockContractViewer(), testutils.NewMockICAControllerKeeper(), testutils.NewMockChannelKeeper()
	cwicaKeeper.SetWasmKeeper(wmKeeper)
	cwicaKeeper.SetICAControllerKeeper(icaCtrlKeeper)
	cwicaKeeper.SetChannelKeeper(channelKeeper)
	contractAddress := e2eTesting.GenContractAddresses(1)[0]
	contractAdminAcc := s.chain.GetAccount(0)
	goCtx := sdk.WrapSDKContext(ctx)

	// TEST CASE 1: invalid msg
	resp, err := cwicaKeeper.SendTx(goCtx, nil)
	s.Require().Nil(resp)
	s.Require().ErrorContains(err, "nil msg is prohibited")

	// TEST CASE 2: empty msg
	resp, err = cwicaKeeper.SendTx(goCtx, &types.MsgSendTx{})
	s.Require().Nil(resp)
	s.Require().ErrorContains(err, "empty Msgs field is prohibited")

	// TEST CASE 3: invalid from address
	cosmosMsg := codectypes.Any{
		TypeUrl: "/cosmos.staking.v1beta1.MsgDelegate",
		Value:   []byte{26, 10, 10, 5, 115, 116, 97, 107, 101, 18, 1, 48},
	}
	resp, err = cwicaKeeper.SendTx(goCtx, &types.MsgSendTx{Msgs: []*codectypes.Any{&cosmosMsg}})
	s.Require().Nil(resp)
	s.Require().ErrorContains(err, "failed to parse address")

	// TEST CASE 4: contract address is not a registered contract
	submitMsg := types.MsgSendTx{
		ContractAddress: contractAddress.String(),
		ConnectionId:    "connection-0",
		Msgs:            []*codectypes.Any{&cosmosMsg},
		Memo:            "memo",
		Timeout:         100,
	}
	s.Require().False(wmKeeper.HasContractInfo(ctx, contractAddress))
	resp, err = cwicaKeeper.SendTx(goCtx, &submitMsg)
	s.Require().Nil(resp)
	s.Require().ErrorContains(err, "is not a contract address")

	// TEST CASE 5: more msgs in the MsgSendTx than is allowed
	wmKeeper.AddContractAdmin(
		contractAddress.String(),
		contractAdminAcc.Address.String(),
	)
	params, err := cwicaKeeper.GetParams(ctx)
	s.Require().NoError(err)
	maxMsgs := params.GetMsgSendTxMaxMessages()
	submitMsg.Msgs = make([]*codectypes.Any, maxMsgs+1)
	s.Require().True(wmKeeper.HasContractInfo(ctx, contractAddress))
	resp, err = cwicaKeeper.SendTx(goCtx, &submitMsg)
	s.Require().Nil(resp)
	s.Require().ErrorContains(err, "MsgSubmitTx contains more messages than allowed")

	// TEST CASE 6: failed to GetActiveChannelID for port
	submitMsg.Msgs = []*codectypes.Any{&cosmosMsg}
	portID := "icacontroller-" + contractAddress.String() + ".ica0"
	resp, err = cwicaKeeper.SendTx(goCtx, &submitMsg)
	s.Require().Nil(resp)
	s.Require().ErrorContains(err, "failed to GetActiveChannelID for port")

	// TEST CASE 7: sequence number not found
	activeChannel := "channel-0"
	icaCtrlKeeper.SetTestStateGetActiveChannelID(activeChannel)
	seq, found := channelKeeper.GetNextSequenceSend(ctx, portID, activeChannel)
	s.Require().False(found)
	s.Require().Equal(uint64(0), seq)
	resp, err = cwicaKeeper.SendTx(goCtx, &submitMsg)
	s.Require().Nil(resp)
	s.Require().ErrorContains(err, "sequence send not found")

	// TEST CASE 8: failed to SendTx as invalid packet sequence
	sequence := uint64(100)
	channelKeeper.SetTestStateNextSequenceSend(sequence)
	seq, found = channelKeeper.GetNextSequenceSend(ctx, portID, activeChannel)
	s.Require().True(found)
	s.Require().Equal(sequence, seq)
	resp, err = cwicaKeeper.SendTx(goCtx, &submitMsg)
	s.Require().Nil(resp)
	s.Require().ErrorContains(err, "failed to SendTx")

	// TEST CASE 9: successfully SendTx
	icaCtrlKeeper.SetTestStateSendTx(100)
	resp, err = cwicaKeeper.SendTx(goCtx, &submitMsg)
	s.Require().Equal(types.MsgSendTxResponse{
		SequenceId: sequence,
		Channel:    activeChannel,
	}, *resp)
	s.Require().NoError(err)
}
