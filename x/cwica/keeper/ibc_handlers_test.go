package keeper_test

import (
	"encoding/json"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/cwica/types"
)

func (s *KeeperTestSuite) TestHandleAcknowledgement() {
	ctx, cwicaKeeper := s.chain.GetContext().WithBlockHeight(100).WithGasMeter(sdk.NewGasMeter(1_000_000_000_000)), s.chain.GetApp().Keepers.CWICAKeeper
	wmKeeper, icaCtrlKeeper, channelKeeper := testutils.NewMockContractViewer(), testutils.NewMockICAControllerKeeper(), testutils.NewMockChannelKeeper()
	cwicaKeeper.SetWasmKeeper(wmKeeper)
	cwicaKeeper.SetICAControllerKeeper(icaCtrlKeeper)
	cwicaKeeper.SetChannelKeeper(channelKeeper)
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
		SourcePort:    icatypes.ControllerPortPrefix + contractAddress.String(),
		SourceChannel: "channel-0",
	}
	relayerAddress := s.chain.GetAccount(1).Address

	err = cwicaKeeper.HandleAcknowledgement(ctx, channeltypes.Packet{}, nil, relayerAddress)
	s.Require().ErrorContains(err, "failed to parse contract address: : invalid address")

	err = cwicaKeeper.HandleAcknowledgement(ctx, p, nil, relayerAddress)
	s.Require().ErrorContains(err, "cannot unmarshal ICS-27 packet acknowledgement")

	sudoMsg := types.SudoPayload{
		ICA: &types.MessageICASuccess{
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
	err = cwicaKeeper.HandleAcknowledgement(ctx, p, resAckData, relayerAddress)
	s.Require().NoError(err)

	// error contract SudoResponse
	wmKeeper.SetReturnSudoError(errors.New("error sudoResponse"))
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	_, err = wmKeeper.Sudo(ctx, contractAddress, msgAck)
	s.Require().ErrorContains(err, "error sudoResponse")
	err = cwicaKeeper.HandleAcknowledgement(ctx, p, resAckData, relayerAddress)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestHandleTimeout() {
	ctx, cwicaKeeper := s.chain.GetContext().WithBlockHeight(100).WithGasMeter(sdk.NewGasMeter(1_000_000_000_000)), s.chain.GetApp().Keepers.CWICAKeeper
	wmKeeper, icaCtrlKeeper, channelKeeper := testutils.NewMockContractViewer(), testutils.NewMockICAControllerKeeper(), testutils.NewMockChannelKeeper()
	cwicaKeeper.SetWasmKeeper(wmKeeper)
	cwicaKeeper.SetICAControllerKeeper(icaCtrlKeeper)
	cwicaKeeper.SetChannelKeeper(channelKeeper)
	contractAddress := e2eTesting.GenContractAddresses(1)[0]
	contractAdminAcc := s.chain.GetAccount(0)
	wmKeeper.AddContractAdmin(
		contractAddress.String(),
		contractAdminAcc.Address.String(),
	)
	relayerAddress := s.chain.GetAccount(1).Address
	p := channeltypes.Packet{
		Sequence:      100,
		SourcePort:    icatypes.ControllerPortPrefix + contractAddress.String(),
		SourceChannel: "channel-0",
	}
	pJson, err := json.Marshal(p)
	s.Require().NoError(err)

	sudoMsg := types.SudoPayload{
		Error: types.NewSudoErrorMsg(types.SudoError{
			ErrorCode: types.ModuleErrors_ERR_PACKET_TIMEOUT,
			Payload:   string(pJson),
		}),
	}
	msgAck, err := json.Marshal(sudoMsg)
	s.Require().NoError(err)

	err = cwicaKeeper.HandleTimeout(ctx, channeltypes.Packet{}, relayerAddress)
	s.Require().ErrorContains(err, "failed to parse contract address: : invalid address")

	// contract success
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	_, err = wmKeeper.Sudo(ctx, contractAddress, msgAck)
	s.Require().NoError(err)
	err = cwicaKeeper.HandleTimeout(ctx, p, relayerAddress)
	s.Require().NoError(err)

	// contract error
	wmKeeper.SetReturnSudoError(errors.New("SudoTimeout error"))
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	_, err = wmKeeper.Sudo(ctx, contractAddress, msgAck)
	s.Require().ErrorContains(err, "SudoTimeout error")
	err = cwicaKeeper.HandleTimeout(ctx, p, relayerAddress)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestHandleChanOpenAck() {
	ctx, cwicaKeeper := s.chain.GetContext().WithBlockHeight(100).WithGasMeter(sdk.NewGasMeter(1_000_000_000_000)), s.chain.GetApp().Keepers.CWICAKeeper
	wmKeeper, icaCtrlKeeper, channelKeeper := testutils.NewMockContractViewer(), testutils.NewMockICAControllerKeeper(), testutils.NewMockChannelKeeper()
	cwicaKeeper.SetWasmKeeper(wmKeeper)
	cwicaKeeper.SetICAControllerKeeper(icaCtrlKeeper)
	cwicaKeeper.SetChannelKeeper(channelKeeper)
	contractAddress := e2eTesting.GenContractAddresses(1)[0]
	contractAdminAcc := s.chain.GetAccount(0)
	wmKeeper.AddContractAdmin(
		contractAddress.String(),
		contractAdminAcc.Address.String(),
	)
	portID := icatypes.ControllerPortPrefix + contractAddress.String()
	channelID := "channel-0"
	counterpartyChannelID := "channel-1"

	err := cwicaKeeper.HandleChanOpenAck(ctx, "", channelID, counterpartyChannelID, "1")
	s.Require().ErrorContains(err, "failed to parse contract address: : invalid address")

	sudoMsg := types.SudoPayload{
		ICA: &types.MessageICASuccess{
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
	err = cwicaKeeper.HandleChanOpenAck(ctx, portID, channelID, counterpartyChannelID, "1")
	s.Require().NoError(err)

	// sudo success
	wmKeeper.SetReturnSudoError(nil)
	_, err = wmKeeper.Sudo(ctx, contractAddress, msg)
	s.Require().NoError(err)
	err = cwicaKeeper.HandleChanOpenAck(ctx, portID, channelID, counterpartyChannelID, "1")
	s.Require().NoError(err)
}
