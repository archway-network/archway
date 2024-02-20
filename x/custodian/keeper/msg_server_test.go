package keeper_test

import (
	"time"

	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/custodian/keeper"
	"github.com/archway-network/archway/x/custodian/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
)

func (s *KeeperTestSuite) TestRegisterInterchainAccount() {
	ctx, custodianKeeper := s.chain.GetContext().WithBlockHeight(100), s.chain.GetApp().Keepers.CustodianKeeper
	wmKeeper, icaCtrlKeeper := testutils.NewMockContractViewer(), testutils.NewMockICAControllerKeeper()
	custodianKeeper.SetWasmKeeper(wmKeeper)
	custodianKeeper.SetICAControllerKeeper(icaCtrlKeeper)
	contractAddress := e2eTesting.GenContractAddresses(1)[0]
	contractAdminAcc := s.chain.GetAccount(0)
	goCtx := sdk.WrapSDKContext(ctx)

	msgRegAcc := types.MsgRegisterInterchainAccount{
		FromAddress:         contractAddress.String(),
		ConnectionId:        "connection-0",
		InterchainAccountId: "ica0",
	}
	icaOwner := types.NewICAOwnerFromAddress(contractAddress, msgRegAcc.InterchainAccountId)

	resp, err := custodianKeeper.RegisterInterchainAccount(goCtx, &types.MsgRegisterInterchainAccount{})
	s.Require().ErrorContains(err, "failed to parse address")
	s.Require().Nil(resp)

	s.Require().False(wmKeeper.HasContractInfo(ctx, contractAddress))
	resp, err = custodianKeeper.RegisterInterchainAccount(goCtx, &msgRegAcc)
	s.Require().ErrorContains(err, "is not a contract address")
	s.Require().Nil(resp)

	wmKeeper.AddContractAdmin(
		contractAddress.String(),
		contractAdminAcc.Address.String(),
	)

	s.Require().True(wmKeeper.HasContractInfo(ctx, contractAddress))
	contractInfo := wmKeeper.GetContractInfo(ctx, contractAddress)
	s.Require().Equal(contractAdminAcc.Address.String(), contractInfo.Admin)

	err = icaCtrlKeeper.RegisterInterchainAccount(ctx, msgRegAcc.ConnectionId, icaOwner.String(), "")
	s.Require().Error(err)
	resp, err = custodianKeeper.RegisterInterchainAccount(goCtx, &msgRegAcc)
	s.Require().ErrorContains(err, "failed to RegisterInterchainAccount")
	s.Require().Nil(resp)

	icaCtrlKeeper.SetTestStateRegisterInterchainAccount(false)
	err = icaCtrlKeeper.RegisterInterchainAccount(ctx, msgRegAcc.ConnectionId, icaOwner.String(), "")
	s.Require().NoError(err)
	resp, err = custodianKeeper.RegisterInterchainAccount(goCtx, &msgRegAcc)
	s.Require().NoError(err)
	s.Require().Equal(types.MsgRegisterInterchainAccountResponse{}, *resp)
}

func (s *KeeperTestSuite) TestSubmitTx() {
	ctx, custodianKeeper := s.chain.GetContext().WithBlockHeight(100), s.chain.GetApp().Keepers.CustodianKeeper
	wmKeeper, icaCtrlKeeper, channelKeeper := testutils.NewMockContractViewer(), testutils.NewMockICAControllerKeeper(), testutils.NewMockChannelKeeper()
	custodianKeeper.SetWasmKeeper(wmKeeper)
	custodianKeeper.SetICAControllerKeeper(icaCtrlKeeper)
	custodianKeeper.SetChannelKeeper(channelKeeper)
	contractAddress := e2eTesting.GenContractAddresses(1)[0]
	contractAdminAcc := s.chain.GetAccount(0)
	goCtx := sdk.WrapSDKContext(ctx)

	cosmosMsg := codectypes.Any{
		TypeUrl: "/cosmos.staking.v1beta1.MsgDelegate",
		Value:   []byte{26, 10, 10, 5, 115, 116, 97, 107, 101, 18, 1, 48},
	}
	submitMsg := types.MsgSubmitTx{
		FromAddress:         contractAddress.String(),
		InterchainAccountId: "ica0",
		ConnectionId:        "connection-0",
		Msgs:                []*codectypes.Any{&cosmosMsg},
		Memo:                "memo",
		Timeout:             100,
	}

	resp, err := custodianKeeper.SubmitTx(goCtx, nil)
	s.Require().Nil(resp)
	s.Require().ErrorContains(err, "nil msg is prohibited")

	resp, err = custodianKeeper.SubmitTx(goCtx, &types.MsgSubmitTx{})
	s.Require().Nil(resp)
	s.Require().ErrorContains(err, "empty Msgs field is prohibited")

	resp, err = custodianKeeper.SubmitTx(goCtx, &types.MsgSubmitTx{Msgs: []*codectypes.Any{&cosmosMsg}})
	s.Require().Nil(resp)
	s.Require().ErrorContains(err, "failed to parse address")

	s.Require().False(wmKeeper.HasContractInfo(ctx, contractAddress))
	resp, err = custodianKeeper.SubmitTx(goCtx, &submitMsg)
	s.Require().Nil(resp)
	s.Require().ErrorContains(err, "is not a contract address")

	wmKeeper.AddContractAdmin(
		contractAddress.String(),
		contractAdminAcc.Address.String(),
	)
	params := custodianKeeper.GetParams(ctx)
	maxMsgs := params.GetMsgSubmitTxMaxMessages()
	submitMsg.Msgs = make([]*codectypes.Any, maxMsgs+1)
	s.Require().True(wmKeeper.HasContractInfo(ctx, contractAddress))
	resp, err = custodianKeeper.SubmitTx(goCtx, &submitMsg)
	s.Require().Nil(resp)
	s.Require().ErrorContains(err, "MsgSubmitTx contains more messages than allowed")
	submitMsg.Msgs = []*codectypes.Any{&cosmosMsg}

	portID := "icacontroller-" + contractAddress.String() + ".ica0"
	cid, found := icaCtrlKeeper.GetActiveChannelID(ctx, "connection-0", portID)
	s.Require().False(found)
	s.Require().Equal("", cid)
	resp, err = custodianKeeper.SubmitTx(goCtx, &submitMsg)
	s.Require().Nil(resp)
	s.Require().ErrorContains(err, "failed to GetActiveChannelID for port")

	activeChannel := "channel-0"
	icaCtrlKeeper.SetTestStateGetActiveChannelID(activeChannel)

	cid, found = icaCtrlKeeper.GetActiveChannelID(ctx, "connection-0", portID)
	s.Require().True(found)
	s.Require().Equal(activeChannel, cid)
	seq, found := channelKeeper.GetNextSequenceSend(ctx, portID, activeChannel)
	s.Require().False(found)
	s.Require().Equal(uint64(0), seq)
	resp, err = custodianKeeper.SubmitTx(goCtx, &submitMsg)
	s.Require().Nil(resp)
	s.Require().ErrorContains(err, "sequence send not found")

	sequence := uint64(100)
	channelKeeper.SetTestStateNextSequenceSend(sequence)
	seq, found = channelKeeper.GetNextSequenceSend(ctx, portID, activeChannel)
	s.Require().True(found)
	s.Require().Equal(sequence, seq)

	data, err := keeper.SerializeCosmosTx(custodianKeeper.Codec, submitMsg.Msgs)
	s.Require().NoError(err)
	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
		Memo: submitMsg.Memo,
	}

	timeoutTimestamp := ctx.BlockTime().Add(time.Duration(submitMsg.Timeout) * time.Second).UnixNano()
	packetSeq, err := icaCtrlKeeper.SendTx(ctx, nil, "connection-0", portID, packetData, uint64(timeoutTimestamp))
	s.Require().Equal(uint64(0), packetSeq)
	s.Require().ErrorContains(err, "failed to send tx")
	resp, err = custodianKeeper.SubmitTx(goCtx, &submitMsg)
	s.Require().Nil(resp)
	s.Require().ErrorContains(err, "failed to SendTx")

	icaCtrlKeeper.SetTestStateSendTx(100)
	packetSeq, err = icaCtrlKeeper.SendTx(ctx, nil, "connection-0", portID, packetData, uint64(timeoutTimestamp))
	s.Require().Equal(uint64(100), packetSeq)
	s.Require().NoError(err)
	resp, err = custodianKeeper.SubmitTx(goCtx, &submitMsg)
	s.Require().Equal(types.MsgSubmitTxResponse{
		SequenceId: sequence,
		Channel:    activeChannel,
	}, *resp)
	s.Require().NoError(err)
}
