package types_test

import (
	"testing"

	cosmosTypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/archway-network/archway/x/cwica/types"
)

const TestAddress = "cosmos1c4k24jzduc365kywrsvf5ujz4ya6mwymy8vq4q"

// TestMsgRegisterInterchainAccountValidate tests the validation of the MsgRegisterInterchainAccount
func TestMsgRegisterInterchainAccountValidate(t *testing.T) {
	tests := []struct {
		name        string
		malleate    func() sdktypes.Msg
		expectedErr error
	}{
		{
			"valid",
			func() sdktypes.Msg {
				return &types.MsgRegisterInterchainAccount{
					ContractAddress: TestAddress,
					ConnectionId:    "connection-id",
				}
			},
			nil,
		},
		{
			"empty ContractAddress",
			func() sdktypes.Msg {
				return &types.MsgRegisterInterchainAccount{
					ContractAddress: "",
					ConnectionId:    "connection-id",
				}
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid ContractAddress",
			func() sdktypes.Msg {
				return &types.MsgRegisterInterchainAccount{
					ContractAddress: "invalid address",
					ConnectionId:    "connection-id",
				}
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"empty connection id",
			func() sdktypes.Msg {
				return &types.MsgRegisterInterchainAccount{
					ContractAddress: TestAddress,
					ConnectionId:    "",
				}
			},
			types.ErrEmptyConnectionID,
		},
	}

	for _, tt := range tests {
		msg := tt.malleate()

		if tt.expectedErr != nil {
			require.ErrorIs(t, msg.ValidateBasic(), tt.expectedErr)
		} else {
			require.NoError(t, msg.ValidateBasic())
		}
	}
}

// TestMsgRegisterInterchainAccountGetSigners tests the GetSigners of the MsgRegisterInterchainAccount
func TestMsgRegisterInterchainAccountGetSigners(t *testing.T) {
	tests := []struct {
		name     string
		malleate func() sdktypes.Msg
		isValid  bool
	}{
		{
			"invalid_signer",
			func() sdktypes.Msg {
				return &types.MsgRegisterInterchainAccount{
					ContractAddress: "👻",
					ConnectionId:    "connection-id",
				}
			},
			false,
		},
		{
			"valid_signer",
			func() sdktypes.Msg {
				return &types.MsgRegisterInterchainAccount{
					ContractAddress: TestAddress,
					ConnectionId:    "connection-id",
				}
			},
			true,
		},
	}

	for _, tt := range tests {
		msg := tt.malleate()
		if tt.isValid {
			addr, _ := sdktypes.AccAddressFromBech32(TestAddress)
			require.Equal(t, msg.GetSigners(), []sdktypes.AccAddress{addr})
		} else {
			require.Panics(t, func() { msg.GetSigners() })
		}
	}
}

// TestMsgSendTxValidate tests the validation of the MsgSendTx
func TestMsgSendTxValidate(t *testing.T) {
	tests := []struct {
		name        string
		malleate    func() sdktypes.Msg
		expectedErr error
	}{
		{
			"valid",
			func() sdktypes.Msg {
				return &types.MsgSendTx{
					ContractAddress: TestAddress,
					ConnectionId:    "connection-id",
					Msgs: []*cosmosTypes.Any{{
						TypeUrl: "msg",
						Value:   []byte{100}, // just check that values are not nil
					}},
					Timeout: 1,
				}
			},
			nil,
		},
		{
			"invalid timeout",
			func() sdktypes.Msg {
				return &types.MsgSendTx{
					ContractAddress: TestAddress,
					ConnectionId:    "connection-id",
					Msgs: []*cosmosTypes.Any{{
						TypeUrl: "msg",
						Value:   []byte{100}, // just check that values are not nil
					}},
					Timeout: 0,
				}
			},
			types.ErrInvalidTimeout,
		},
		{
			"empty connection id",
			func() sdktypes.Msg {
				return &types.MsgSendTx{
					ContractAddress: TestAddress,
					ConnectionId:    "",
					Msgs: []*cosmosTypes.Any{{
						TypeUrl: "msg",
						Value:   []byte{100}, // just check that values are not nil
					}},
					Timeout: 1,
				}
			},
			types.ErrEmptyConnectionID,
		},
		{
			"no messages",
			func() sdktypes.Msg {
				return &types.MsgSendTx{
					ContractAddress: TestAddress,
					ConnectionId:    "connection-id",
					Msgs:            nil,
					Timeout:         1,
				}
			},
			types.ErrNoMessages,
		},
		{
			"empty ContractAddress",
			func() sdktypes.Msg {
				return &types.MsgSendTx{
					ContractAddress: "",
					ConnectionId:    "connection-id",
					Msgs: []*cosmosTypes.Any{{
						TypeUrl: "msg",
						Value:   []byte{100}, // just check that values are not nil
					}},
					Timeout: 1,
				}
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid ContractAddress",
			func() sdktypes.Msg {
				return &types.MsgSendTx{
					ContractAddress: "👻",
					ConnectionId:    "connection-id",
					Msgs: []*cosmosTypes.Any{{
						TypeUrl: "msg",
						Value:   []byte{100}, // just check that values are not nil
					}},
					Timeout: 1,
				}
			},
			sdkerrors.ErrInvalidAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.malleate()

			if tt.expectedErr != nil {
				require.ErrorIs(t, msg.ValidateBasic(), tt.expectedErr)
			} else {
				require.NoError(t, msg.ValidateBasic())
			}
		})
	}
}

func TestMsgSubmitTXGetSigners(t *testing.T) {
	tests := []struct {
		name     string
		malleate func() sdktypes.Msg
	}{
		{
			"valid_signer",
			func() sdktypes.Msg {
				return &types.MsgSendTx{
					ContractAddress: TestAddress,
					ConnectionId:    "connection-id",
					Msgs: []*cosmosTypes.Any{{
						TypeUrl: "msg",
						Value:   []byte{100}, // just check that values are not nil
					}},
				}
			},
		},
	}

	for _, tt := range tests {
		msg := tt.malleate()
		addr, _ := sdktypes.AccAddressFromBech32(TestAddress)
		require.Equal(t, msg.GetSigners(), []sdktypes.AccAddress{addr})
	}
}
