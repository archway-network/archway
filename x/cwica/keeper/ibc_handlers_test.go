package keeper_test

import (
	"errors"

	storetypes "cosmossdk.io/store/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
)

// TestKeeper_HandleChanOpenAck tests the HandleChanOpenAck method
func (s *KeeperTestSuite) TestHandleAcknowledgement() {
	ctx, cwicaKeeper := s.chain.GetContext().WithBlockHeight(100).WithGasMeter(storetypes.NewGasMeter(1_000_000_000_000)), s.chain.GetApp().Keepers.CWICAKeeper
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

	// TEST CASE 1: invalid contract address
	err := cwicaKeeper.HandleAcknowledgement(ctx, channeltypes.Packet{}, nil)
	s.Require().ErrorContains(err, "failed to parse contract address: : invalid address")

	// TEST CASE 2: invalid packet acknowledgement
	p := channeltypes.Packet{
		Sequence:      100,
		SourcePort:    icatypes.ControllerPortPrefix + contractAddress.String(),
		SourceChannel: "channel-0",
	}
	err = cwicaKeeper.HandleAcknowledgement(ctx, p, nil)
	s.Require().ErrorContains(err, "cannot unmarshal ICS-27 packet acknowledgement")

	// TEST CASE 3: success contract SudoResponse
	resACK := channeltypes.Acknowledgement{
		Response: &channeltypes.Acknowledgement_Result{Result: []byte("Result")},
	}
	resAckData, err := channeltypes.SubModuleCdc.MarshalJSON(&resACK)
	s.Require().NoError(err)

	err = cwicaKeeper.HandleAcknowledgement(ctx, p, resAckData)
	s.Require().NoError(err)

	// TEST CASE 4: contract callback fails - should not return error - because error is swallowed
	wmKeeper.SetReturnSudoError(errors.New("error sudoResponse"))
	err = cwicaKeeper.HandleAcknowledgement(ctx, p, resAckData)
	s.Require().NoError(err)
}

// TestKeeper_HandleChanOpenAck tests the HandleChanOpenAck method
func (s *KeeperTestSuite) TestHandleTimeout() {
	ctx, cwicaKeeper := s.chain.GetContext().WithBlockHeight(100).WithGasMeter(storetypes.NewGasMeter(1_000_000_000_000)), s.chain.GetApp().Keepers.CWICAKeeper
	wmKeeper, icaCtrlKeeper, channelKeeper := testutils.NewMockContractViewer(), testutils.NewMockICAControllerKeeper(), testutils.NewMockChannelKeeper()
	errorsKeeper := s.chain.GetApp().Keepers.CWErrorsKeeper
	errorsKeeper.SetWasmKeeper(wmKeeper)
	cwicaKeeper.SetWasmKeeper(wmKeeper)
	cwicaKeeper.SetICAControllerKeeper(icaCtrlKeeper)
	cwicaKeeper.SetChannelKeeper(channelKeeper)
	cwicaKeeper.SetErrorsKeeper(errorsKeeper)
	contractAddress := e2eTesting.GenContractAddresses(1)[0]
	contractAdminAcc := s.chain.GetAccount(0)
	wmKeeper.AddContractAdmin(
		contractAddress.String(),
		contractAdminAcc.Address.String(),
	)

	// TEST CASE 1: invalid contract address
	err := cwicaKeeper.HandleTimeout(ctx, channeltypes.Packet{})
	s.Require().ErrorContains(err, "failed to parse contract address: : invalid address")

	// TEST CASE 2: success contract SudoResponse
	p := channeltypes.Packet{
		Sequence:      100,
		SourcePort:    icatypes.ControllerPortPrefix + contractAddress.String(),
		SourceChannel: "channel-0",
	}

	err = cwicaKeeper.HandleTimeout(ctx, p)
	s.Require().NoError(err)

	// TEST CASE 3: contract callback fails - should not return error - because error is swallowed
	wmKeeper.SetReturnSudoError(errors.New("SudoTimeout error"))
	err = cwicaKeeper.HandleTimeout(ctx, p)
	s.Require().NoError(err)
}

// TestKeeper_HandleChanOpenAck tests the HandleChanOpenAck method
func (s *KeeperTestSuite) TestHandleChanOpenAck() {
	ctx, cwicaKeeper := s.chain.GetContext().WithBlockHeight(100).WithGasMeter(storetypes.NewGasMeter(1_000_000_000_000)), s.chain.GetApp().Keepers.CWICAKeeper
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

	// TEST CASE 1: invalid contract address
	err := cwicaKeeper.HandleChanOpenAck(ctx, "", channelID, counterpartyChannelID, "1")
	s.Require().ErrorContains(err, "failed to parse contract address: : invalid address")

	icaMetadata := icatypes.Metadata{
		Version:                "ics27-1",
		ControllerConnectionId: "connection-0",
		HostConnectionId:       "connection-0",
		Encoding:               icatypes.EncodingProtobuf,
		TxType:                 icatypes.TxTypeSDKMultiMsg,
	}
	icaMetadataBytes, err := icatypes.ModuleCdc.MarshalJSON(&icaMetadata)
	s.Require().NoError(err)

	// TEST CASE 2: success contract SudoResponse
	err = cwicaKeeper.HandleChanOpenAck(ctx, portID, channelID, counterpartyChannelID, string(icaMetadataBytes))
	s.Require().NoError(err)

	// TEST CASE 3: contract callback fails - should not return error - because error is swallowed
	wmKeeper.SetReturnSudoError(errors.New("SudoOnChanOpenAck error"))
	err = cwicaKeeper.HandleChanOpenAck(ctx, portID, channelID, counterpartyChannelID, string(icaMetadataBytes))
	s.Require().NoError(err)
}
