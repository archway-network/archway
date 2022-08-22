package integration

import (
	mocks "github.com/CosmWasm/wasmvm/api"
	wasmVmTypes "github.com/CosmWasm/wasmvm/types"

	"github.com/archway-network/voter/src/state"
	"github.com/archway-network/voter/src/types"
)

func (s *ContractTestSuite) TestIBCChannelOps() {
	env := mocks.MockEnv()

	okChannel := wasmVmTypes.IBCChannel{
		Endpoint: wasmVmTypes.IBCEndpoint{
			ChannelID: "channel-1",
		},
		CounterpartyEndpoint: wasmVmTypes.IBCEndpoint{
			ChannelID: "channel-2",
		},
		Order:   wasmVmTypes.Unordered,
		Version: types.IBCVersion,
	}

	s.Run("ChannelOpen: OpenInit: OK", func() {
		msg := wasmVmTypes.IBCChannelOpenMsg{
			OpenInit: &wasmVmTypes.IBCOpenInit{
				Channel: okChannel,
			},
		}

		_, err := s.instance.IBCChannelOpen(env, msg)
		s.Assert().NoError(err)
	})

	s.Run("ChannelOpen: OpenInit: Fail: invalid channel", func() {
		ch := okChannel
		ch.Endpoint.ChannelID = "invalid"

		msg := wasmVmTypes.IBCChannelOpenMsg{
			OpenInit: &wasmVmTypes.IBCOpenInit{
				Channel: ch,
			},
		}

		_, err := s.instance.IBCChannelOpen(env, msg)
		s.Assert().Error(err)
	})

	s.Run("ChannelOpen: OpenTry: OK", func() {
		msg := wasmVmTypes.IBCChannelOpenMsg{
			OpenTry: &wasmVmTypes.IBCOpenTry{
				Channel:             okChannel,
				CounterpartyVersion: types.IBCVersion,
			},
		}

		_, err := s.instance.IBCChannelOpen(env, msg)
		s.Assert().NoError(err)
	})

	s.Run("ChannelOpen: OpenTry: Fail: invalid counterparty version", func() {
		msg := wasmVmTypes.IBCChannelOpenMsg{
			OpenTry: &wasmVmTypes.IBCOpenTry{
				Channel:             okChannel,
				CounterpartyVersion: "v1.0",
			},
		}

		_, err := s.instance.IBCChannelOpen(env, msg)
		s.Assert().Error(err)
	})

	s.Run("ChannelConnect: OpenConfirm: OK", func() {
		msg := wasmVmTypes.IBCChannelConnectMsg{
			OpenConfirm: &wasmVmTypes.IBCOpenConfirm{},
		}

		_, _, err := s.instance.IBCChannelConnect(env, msg)
		s.Assert().NoError(err)
	})

	s.Run("ChannelConnect: OpenAck: OK", func() {
		msg := wasmVmTypes.IBCChannelConnectMsg{
			OpenAck: &wasmVmTypes.IBCOpenAck{
				CounterpartyVersion: types.IBCVersion,
			},
		}

		_, _, err := s.instance.IBCChannelConnect(env, msg)
		s.Assert().NoError(err)
	})

	s.Run("ChannelConnect: OpenAck: Fail: invalid counterparty version", func() {
		msg := wasmVmTypes.IBCChannelConnectMsg{
			OpenAck: &wasmVmTypes.IBCOpenAck{
				CounterpartyVersion: "v1.0",
			},
		}

		_, _, err := s.instance.IBCChannelConnect(env, msg)
		s.Assert().Error(err)
	})
}

func (s *ContractTestSuite) TestIBCPacketAckTimeout() {
	env := mocks.MockEnv()
	senderAddr := "SenderAddr"

	s.Run("Ack: OK: acked", func() {
		ibcOrigMsg := s.IBCVote(env, senderAddr, 1, "a", "yes", "channel-1")
		ibcOrigMsgBz, err := ibcOrigMsg.MarshalJSON()
		s.Require().NoError(err)

		// Send ack
		msg := wasmVmTypes.IBCPacketAckMsg{
			Acknowledgement: wasmVmTypes.IBCAcknowledgement{
				Data: types.IBCAckDataOK,
			},
			OriginalPacket: wasmVmTypes.IBCPacket{
				Data: ibcOrigMsgBz,
			},
		}

		_, _, err = s.instance.IBCPacketAck(env, msg)
		s.Require().NoError(err)

		ibcStatsRcv := s.GetIBCStats(env, ibcOrigMsg.Vote.From, ibcOrigMsg.Vote.ID)
		s.Assert().Equal(state.IBCPkgAckedStatus, ibcStatsRcv.Status)
	})

	s.Run("Ack: OK: rejected", func() {
		ibcOrigMsg := s.IBCVote(env, senderAddr, 2, "a", "yes", "channel-1")
		ibcOrigMsgBz, err := ibcOrigMsg.MarshalJSON()
		s.Require().NoError(err)

		// Send ack
		msg := wasmVmTypes.IBCPacketAckMsg{
			Acknowledgement: wasmVmTypes.IBCAcknowledgement{},
			OriginalPacket: wasmVmTypes.IBCPacket{
				Data: ibcOrigMsgBz,
			},
		}

		_, _, err = s.instance.IBCPacketAck(env, msg)
		s.Require().NoError(err)

		ibcStatsRcv := s.GetIBCStats(env, ibcOrigMsg.Vote.From, ibcOrigMsg.Vote.ID)
		s.Assert().Equal(state.IBCPkgRejectedStatus, ibcStatsRcv.Status)
	})

	s.Run("Timeout: OK", func() {
		ibcOrigMsg := s.IBCVote(env, senderAddr, 1, "a", "yes", "channel-1")
		ibcOrigMsgBz, err := ibcOrigMsg.MarshalJSON()
		s.Require().NoError(err)

		// Send ack
		msg := wasmVmTypes.IBCPacketTimeoutMsg{
			Packet: wasmVmTypes.IBCPacket{
				Data: ibcOrigMsgBz,
			},
		}

		_, _, err = s.instance.IBCPacketTimeout(env, msg)
		s.Require().NoError(err)

		ibcStatsRcv := s.GetIBCStats(env, ibcOrigMsg.Vote.From, ibcOrigMsg.Vote.ID)
		s.Assert().Equal(state.IBCPkgTimedOutStatus, ibcStatsRcv.Status)
	})
}
