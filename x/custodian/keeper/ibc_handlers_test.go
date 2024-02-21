package keeper_test

import (
	"encoding/json"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/custodian/types"
)

func (s *KeeperTestSuite) TestHandleAcknowledgement() {
	ctx, custodianKeeper := s.chain.GetContext().WithBlockHeight(100).WithGasMeter(sdk.NewGasMeter(1_000_000_000_000)), s.chain.GetApp().Keepers.CustodianKeeper
	wmKeeper, icaCtrlKeeper, channelKeeper := testutils.NewMockContractViewer(), testutils.NewMockICAControllerKeeper(), testutils.NewMockChannelKeeper()
	custodianKeeper.SetWasmKeeper(wmKeeper)
	custodianKeeper.SetICAControllerKeeper(icaCtrlKeeper)
	custodianKeeper.SetChannelKeeper(channelKeeper)
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

	err = custodianKeeper.HandleAcknowledgement(ctx, channeltypes.Packet{}, nil, relayerAddress)
	s.Require().ErrorContains(err, "failed to get ica owner from port")

	err = custodianKeeper.HandleAcknowledgement(ctx, p, nil, relayerAddress)
	s.Require().ErrorContains(err, "cannot unmarshal ICS-27 packet acknowledgement")

	sudoMsg := types.SudoPayload{
		Custodian: &types.MessageCustodianSuccess{
			TxExecuted: &types.ICATxResponse{
				Data:   resACK.GetResult(),
				Packet: p,
			},
		},
	}
	msgAck, err := json.Marshal(sudoMsg)
	s.Require().NoError(err)

	// success contract SudoResponse
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	_, err = wmKeeper.Sudo(ctx, contractAddress, msgAck)
	s.Require().NoError(err)
	err = custodianKeeper.HandleAcknowledgement(ctx, p, resAckData, relayerAddress)
	s.Require().NoError(err)

	// error contract SudoResponse
	wmKeeper.SetReturnSudoError(errors.New("error sudoResponse"))
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	_, err = wmKeeper.Sudo(ctx, contractAddress, msgAck)
	s.Require().ErrorContains(err, "error sudoResponse")
	err = custodianKeeper.HandleAcknowledgement(ctx, p, resAckData, relayerAddress)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestHandleTimeout() {
	ctx, custodianKeeper := s.chain.GetContext().WithBlockHeight(100).WithGasMeter(sdk.NewGasMeter(1_000_000_000_000)), s.chain.GetApp().Keepers.CustodianKeeper
	wmKeeper, icaCtrlKeeper, channelKeeper := testutils.NewMockContractViewer(), testutils.NewMockICAControllerKeeper(), testutils.NewMockChannelKeeper()
	custodianKeeper.SetWasmKeeper(wmKeeper)
	custodianKeeper.SetICAControllerKeeper(icaCtrlKeeper)
	custodianKeeper.SetChannelKeeper(channelKeeper)
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

	sudoMsg := types.SudoPayload{
		Error: &types.MessageCustodianError{
			Timeout: &types.ICATxTimeout{Packet: p},
		},
	}
	msgAck, err := json.Marshal(sudoMsg)
	s.Require().NoError(err)

	err = custodianKeeper.HandleTimeout(ctx, channeltypes.Packet{}, relayerAddress)
	s.Require().ErrorContains(err, "failed to get ica owner from port")

	// contract success
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	_, err = wmKeeper.Sudo(ctx, contractAddress, msgAck)
	s.Require().NoError(err)
	err = custodianKeeper.HandleTimeout(ctx, p, relayerAddress)
	s.Require().NoError(err)

	// contract error
	wmKeeper.SetReturnSudoError(errors.New("SudoTimeout error"))
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	_, err = wmKeeper.Sudo(ctx, contractAddress, msgAck)
	s.Require().ErrorContains(err, "SudoTimeout error")
	err = custodianKeeper.HandleTimeout(ctx, p, relayerAddress)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestHandleChanOpenAck() {
	ctx, custodianKeeper := s.chain.GetContext().WithBlockHeight(100).WithGasMeter(sdk.NewGasMeter(1_000_000_000_000)), s.chain.GetApp().Keepers.CustodianKeeper
	wmKeeper, icaCtrlKeeper, channelKeeper := testutils.NewMockContractViewer(), testutils.NewMockICAControllerKeeper(), testutils.NewMockChannelKeeper()
	custodianKeeper.SetWasmKeeper(wmKeeper)
	custodianKeeper.SetICAControllerKeeper(icaCtrlKeeper)
	custodianKeeper.SetChannelKeeper(channelKeeper)
	contractAddress := e2eTesting.GenContractAddresses(1)[0]
	contractAdminAcc := s.chain.GetAccount(0)
	wmKeeper.AddContractAdmin(
		contractAddress.String(),
		contractAdminAcc.Address.String(),
	)
	portID := icatypes.ControllerPortPrefix + contractAddress.String() + ".ica0"
	channelID := "channel-0"
	counterpartyChannelID := "channel-1"

	err := custodianKeeper.HandleChanOpenAck(ctx, "", channelID, counterpartyChannelID, "1")
	s.Require().ErrorContains(err, "failed to get ica owner from port")

	sudoMsg := types.SudoPayload{
		Custodian: &types.MessageCustodianSuccess{
			AccountRegistered: &types.OpenAckDetails{
				PortID:                portID,
				ChannelID:             channelID,
				CounterpartyChannelID: counterpartyChannelID,
				CounterpartyVersion:   "1",
			},
		},
	}
	msg, err := json.Marshal(sudoMsg)
	s.Require().NoError(err)

	// sudo error
	wmKeeper.SetReturnSudoError(errors.New("SudoOnChanOpenAck error"))
	_, err = wmKeeper.Sudo(ctx, contractAddress, msg)
	s.Require().ErrorContains(err, "SudoOnChanOpenAck error")
	err = custodianKeeper.HandleChanOpenAck(ctx, portID, channelID, counterpartyChannelID, "1")
	s.Require().NoError(err)

	// sudo success
	wmKeeper.SetReturnSudoError(nil)
	_, err = wmKeeper.Sudo(ctx, contractAddress, msg)
	s.Require().NoError(err)
	err = custodianKeeper.HandleChanOpenAck(ctx, portID, channelID, counterpartyChannelID, "1")
	s.Require().NoError(err)
}
