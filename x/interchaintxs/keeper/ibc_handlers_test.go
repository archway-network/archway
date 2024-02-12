package keeper_test

import (
	"errors"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/interchaintxs/keeper"
	"github.com/archway-network/archway/x/interchaintxs/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
)

func (s *KeeperTestSuite) TestHandleAcknowledgement() {
	ctx, icak := s.chain.GetContext().WithBlockHeight(100).WithGasMeter(sdk.NewGasMeter(1_000_000_000_000)), s.chain.GetApp().Keepers.InterchainTxsKeeper
	wmKeeper, icaCtrlKeeper, channelKeeper := testutils.NewMockContractViewer(), testutils.NewMockICAControllerKeeper(), testutils.NewMockChannelKeeper()
	icak.SetWasmKeeper(wmKeeper)
	icak.SetICAControllerKeeper(icaCtrlKeeper)
	icak.SetChannelKeeper(channelKeeper)
	contractAddress := e2eTesting.GenContractAddresses(1)[0]
	contractAdminAcc := s.chain.GetAccount(0)
	wmKeeper.AddContractAdmin(
		contractAddress.String(),
		contractAdminAcc.Address.String(),
	)
	resACK := channeltypes.Acknowledgement{
		Response: &channeltypes.Acknowledgement_Result{Result: []byte("Result")},
	}
	resAckData, err := channeltypes.SubModuleCdc.MarshalJSON(&resACK)
	s.Require().NoError(err)
	p := channeltypes.Packet{
		Sequence:      100,
		SourcePort:    icatypes.ControllerPortPrefix + contractAddress.String() + ".ica0",
		SourceChannel: "channel-0",
	}
	relayerAddress := s.chain.GetAccount(1).Address

	err = icak.HandleAcknowledgement(ctx, channeltypes.Packet{}, nil, relayerAddress)
	s.Require().ErrorContains(err, "failed to get ica owner from port")

	err = icak.HandleAcknowledgement(ctx, p, nil, relayerAddress)
	s.Require().ErrorContains(err, "cannot unmarshal ICS-27 packet acknowledgement")

	msgAck, err := keeper.PrepareSudoCallbackMessage(p, &resACK)
	s.Require().NoError(err)

	// success contract SudoResponse
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	wmKeeper.Sudo(ctx, contractAddress, msgAck)
	err = icak.HandleAcknowledgement(ctx, p, resAckData, relayerAddress)
	s.Require().NoError(err)

	// error contract SudoResponse
	wmKeeper.SetReturnSudoError(errors.New("error sudoResponse"))
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	_, err = wmKeeper.Sudo(ctx, contractAddress, msgAck)
	s.Require().ErrorContains(err, "error sudoResponse")
	err = icak.HandleAcknowledgement(ctx, p, resAckData, relayerAddress)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestHandleTimeout() {
	ctx, icak := s.chain.GetContext().WithBlockHeight(100).WithGasMeter(sdk.NewGasMeter(1_000_000_000_000)), s.chain.GetApp().Keepers.InterchainTxsKeeper
	wmKeeper, icaCtrlKeeper, channelKeeper := testutils.NewMockContractViewer(), testutils.NewMockICAControllerKeeper(), testutils.NewMockChannelKeeper()
	icak.SetWasmKeeper(wmKeeper)
	icak.SetICAControllerKeeper(icaCtrlKeeper)
	icak.SetChannelKeeper(channelKeeper)
	contractAddress := e2eTesting.GenContractAddresses(1)[0]
	contractAdminAcc := s.chain.GetAccount(0)
	wmKeeper.AddContractAdmin(
		contractAddress.String(),
		contractAdminAcc.Address.String(),
	)
	relayerAddress := s.chain.GetAccount(1).Address
	p := channeltypes.Packet{
		Sequence:      100,
		SourcePort:    icatypes.ControllerPortPrefix + contractAddress.String() + ".ica0",
		SourceChannel: "channel-0",
	}

	msgAck, err := keeper.PrepareSudoCallbackMessage(p, nil)
	s.Require().NoError(err)

	err = icak.HandleTimeout(ctx, channeltypes.Packet{}, relayerAddress)
	s.Require().ErrorContains(err, "failed to get ica owner from port")

	// contract success
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	wmKeeper.Sudo(ctx, contractAddress, msgAck)
	err = icak.HandleTimeout(ctx, p, relayerAddress)
	s.Require().NoError(err)

	// contract error
	wmKeeper.SetReturnSudoError(errors.New("SudoTimeout error"))
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	_, err = wmKeeper.Sudo(ctx, contractAddress, msgAck)
	s.Require().ErrorContains(err, "SudoTimeout error")
	err = icak.HandleTimeout(ctx, p, relayerAddress)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestHandleChanOpenAck() {
	ctx, icak := s.chain.GetContext().WithBlockHeight(100).WithGasMeter(sdk.NewGasMeter(1_000_000_000_000)), s.chain.GetApp().Keepers.InterchainTxsKeeper
	wmKeeper, icaCtrlKeeper, channelKeeper := testutils.NewMockContractViewer(), testutils.NewMockICAControllerKeeper(), testutils.NewMockChannelKeeper()
	icak.SetWasmKeeper(wmKeeper)
	icak.SetICAControllerKeeper(icaCtrlKeeper)
	icak.SetChannelKeeper(channelKeeper)
	contractAddress := e2eTesting.GenContractAddresses(1)[0]
	contractAdminAcc := s.chain.GetAccount(0)
	wmKeeper.AddContractAdmin(
		contractAddress.String(),
		contractAdminAcc.Address.String(),
	)
	portID := icatypes.ControllerPortPrefix + contractAddress.String() + ".ica0"
	channelID := "channel-0"
	counterpartyChannelID := "channel-1"

	err := icak.HandleChanOpenAck(ctx, "", channelID, counterpartyChannelID, "1")
	s.Require().ErrorContains(err, "failed to get ica owner from port")

	msg, err := keeper.PrepareOpenAckCallbackMessage(types.OpenAckDetails{
		PortID:                portID,
		ChannelID:             channelID,
		CounterpartyChannelID: counterpartyChannelID,
		CounterpartyVersion:   "1",
	})
	s.Require().NoError(err)

	// sudo error
	wmKeeper.SetReturnSudoError(errors.New("SudoOnChanOpenAck error"))
	_, err = wmKeeper.Sudo(ctx, contractAddress, msg)
	s.Require().ErrorContains(err, "SudoOnChanOpenAck error")
	err = icak.HandleChanOpenAck(ctx, portID, channelID, counterpartyChannelID, "1")
	s.Require().NoError(err)

	// sudo success
	wmKeeper.SetReturnSudoError(nil)
	wmKeeper.Sudo(ctx, contractAddress, msg)
	err = icak.HandleChanOpenAck(ctx, portID, channelID, counterpartyChannelID, "1")
	s.Require().NoError(err)
}
