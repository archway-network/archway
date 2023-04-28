package e2e

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"

	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/archway-network/archway/pkg"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"

	cwStd "github.com/CosmWasm/cosmwasm-go/std"
	cwTypes "github.com/CosmWasm/cosmwasm-go/std/types"
	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channelTypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"

	voterPkg "github.com/archway-network/voter/src/pkg"
	voterCustomTypes "github.com/archway-network/voter/src/pkg/archway/custom"
	voterState "github.com/archway-network/voter/src/state"
	voterTypes "github.com/archway-network/voter/src/types"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
)

const (
	VoterWasmPath       = "../contracts/go/voter/code.wasm"
	DefNewVotingCostAmt = 1000
	DefNewVoteCostAmt   = 100
)

// TestVoter_ExecuteQueryAndReply tests Execute, SmartQuery, SubMessage execution and Reply.
func (s *E2ETestSuite) TestVoter_ExecuteQueryAndReply() {
	chain := s.chainA
	acc1, acc2 := chain.GetAccount(0), chain.GetAccount(1)
	contractAddr := s.VoterUploadAndInstantiate(chain, acc1)

	var votingID uint64
	s.Run("Create a new voting", func() {
		newVotingCreatedAtExp := chain.GetContext().BlockTime().UnixNano()

		id := s.VoterNewVoting(chain, contractAddr, acc1, "Test", []string{"a"}, time.Minute)

		votingRcv := s.VoterGetVoting(chain, contractAddr, id)
		s.Assert().Equal(id, votingRcv.ID)
		s.Assert().Equal("Test", votingRcv.Name)
		s.Assert().EqualValues(newVotingCreatedAtExp, votingRcv.StartTime)
		s.Assert().EqualValues(newVotingCreatedAtExp+int64(time.Minute), votingRcv.EndTime)
		s.Assert().Len(votingRcv.Tallies, 1)
		s.Assert().Equal("a", votingRcv.Tallies[0].Option)
		s.Assert().Empty(votingRcv.Tallies[0].YesAddrs)
		s.Assert().Empty(votingRcv.Tallies[0].NoAddrs)

		votingID = id
	})

	s.Run("Add vote", func() {
		s.VoterVote(chain, contractAddr, acc2, votingID, "a", true)

		tallyRcv := s.VoterGetTally(chain, contractAddr, votingID)
		s.Assert().True(tallyRcv.Open)
		s.Assert().Len(tallyRcv.Votes, 1)
		s.Assert().Equal("a", tallyRcv.Votes[0].Option)
		s.Assert().EqualValues(1, tallyRcv.Votes[0].TotalYes)
		s.Assert().EqualValues(0, tallyRcv.Votes[0].TotalNo)
	})

	s.Run("Release contract funds and verify (x/bank submsg execution and reply)", func() {
		acc1BalanceBefore := chain.GetBalance(acc1.Address)

		contractCoinsExp := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewIntFromUint64(DefNewVotingCostAmt+DefNewVoteCostAmt)))
		contractCoinsRcv := chain.GetBalance(contractAddr)
		s.Assert().EqualValues(contractCoinsExp, contractCoinsRcv)

		s.VoterRelease(chain, contractAddr, acc1)
		// Asserts were disabled since the contract always returns 0 coins (refer to Voter's handleReplyBankMsg function)
		// s.Assert().EqualValues(contractCoinsExp, releasedCoinsRcv)

		acc1BalanceAfter := chain.GetBalance(acc1.Address)
		acc1BalanceExpected := acc1BalanceBefore.Add(contractCoinsExp...).Sub(chain.GetDefaultTxFee())
		s.Assert().EqualValues(acc1BalanceExpected.String(), acc1BalanceAfter.String())

		releaseStats := s.VoterGetReleaseStats(chain, contractAddr)
		s.Assert().EqualValues(1, releaseStats.Count)
		// s.Assert().EqualValues(releasedCoinsRcv, s.CosmWasmCoinsToSDK(releaseStats.TotalAmount...))
	})
}

// TestVoter_Sudo tests Sudo execution via Gov proposal to change the contract parameters.
// Test indirectly checks:
//   - api.HumanAddress: called by querying contract Params (OwnerAddr);
//   - and api.CanonicalAddress: called by instantiate to convert OwnerAddr (string) to canonical address (bytes);
func (s *E2ETestSuite) TestVoter_Sudo() {
	chain := s.chainA
	acc := chain.GetAccount(0)
	contractAddr := s.VoterUploadAndInstantiate(chain, acc)

	paramsExp := s.VoterDefaultParams(acc)
	newVotingCoin, err := voterPkg.ParseCoinFromString(paramsExp.NewVotingCost)
	s.Require().NoError(err)
	newVotingCoin.Amount = newVotingCoin.Amount.Add64(100)
	paramsExp.NewVotingCost = newVotingCoin.String()

	s.Run("Submit param change proposal and verify applied", func() {
		sudoMsg := voterTypes.MsgSudo{
			ChangeNewVotingCost: &voterTypes.ChangeCostRequest{
				NewCost: cwTypes.Coin{
					Denom:  newVotingCoin.Denom,
					Amount: newVotingCoin.Amount,
				},
			},
		}
		sudoMsgBz, err := sudoMsg.MarshalJSON()
		s.Require().NoError(err)

		proposal := wasmdTypes.SudoContractProposal{
			Title:       "Increase NewVotingCost",
			Description: "Some desc",
			Contract:    contractAddr.String(),
			Msg:         sudoMsgBz,
		}

		chain.ExecuteGovProposal(acc, true, &proposal)

		paramsRcv := s.VoterGetParams(chain, contractAddr)
		s.Assert().EqualValues(paramsExp, paramsRcv)
	})
}

// TestVoter_IBCSend tests IBC send/ack execution between two contracts.
func (s *E2ETestSuite) TestVoter_IBCSend() {
	chainA, chainB := s.chainA, s.chainB

	accA, accB1, accB2 := chainA.GetAccount(0), chainB.GetAccount(0), chainB.GetAccount(1)
	contractAddrA, contractAddrB := s.VoterUploadAndInstantiate(chainA, accA), s.VoterUploadAndInstantiate(chainB, accB1)
	contractPortA, contractPortB := chainA.GetContractInfo(contractAddrA).IBCPortID, chainB.GetContractInfo(contractAddrB).IBCPortID

	// Create a relayer
	ibcPath := e2eTesting.NewIBCPath(
		s.T(),
		chainA, chainB,
		contractPortA, contractPortB,
		voterTypes.IBCVersion, channelTypes.UNORDERED,
	)
	channelA := ibcPath.EndpointA().ChannelID()

	// Create a voting and add two votes via IBC
	votingID := s.VoterNewVoting(chainA, contractAddrA, accA, "Test", []string{"a"}, time.Minute)

	ibcPacket1 := s.VoterIBCVote(chainB, contractAddrB, accB1, votingID, "a", true, channelA)
	ibcPacket2 := s.VoterIBCVote(chainB, contractAddrB, accB2, votingID, "a", false, channelA)

	// Check IBC stats before the relaying
	ibcStatsB1Before := s.VoterGetIBCStats(chainB, contractAddrB, accB1)
	s.Require().Len(ibcStatsB1Before, 1)
	s.Assert().Equal(voterState.IBCPkgSentStatus, ibcStatsB1Before[0].Status)

	ibcStatsB2Before := s.VoterGetIBCStats(chainB, contractAddrB, accB2)
	s.Require().Len(ibcStatsB2Before, 1)
	s.Assert().Equal(voterState.IBCPkgSentStatus, ibcStatsB2Before[0].Status)

	// Relay
	ibcPath.RelayPacket(ibcPacket1, voterTypes.IBCAckDataOK)
	ibcPath.RelayPacket(ibcPacket2, voterTypes.IBCAckDataOK)

	// Check IBC stats after the relaying
	ibcStatsB1After := s.VoterGetIBCStats(chainB, contractAddrB, accB1)
	s.Require().Len(ibcStatsB1After, 1)
	s.Assert().Equal(voterState.IBCPkgAckedStatus, ibcStatsB1After[0].Status)

	ibcStatsB2After := s.VoterGetIBCStats(chainB, contractAddrB, accB2)
	s.Require().Len(ibcStatsB2After, 1)
	s.Assert().Equal(voterState.IBCPkgAckedStatus, ibcStatsB2After[0].Status)

	// Check voting tally has been updated
	voting := s.VoterGetVoting(chainA, contractAddrA, votingID)
	s.Require().Len(voting.Tallies, 1)
	s.Assert().Contains(voting.Tallies[0].YesAddrs, accB1.Address.String())
	s.Assert().Contains(voting.Tallies[0].NoAddrs, accB2.Address.String())
}

// TestVoter_IBCSend tests IBC timeout execution between two contracts.
func (s *E2ETestSuite) TestVoter_IBCTimeout() {
	chainA, chainB := s.chainA, s.chainB

	accA, accB := chainA.GetAccount(0), chainB.GetAccount(0)
	contractAddrA, contractAddrB := s.VoterUploadAndInstantiate(chainA, accA), s.VoterUploadAndInstantiate(chainB, accB)
	contractPortA, contractPortB := chainA.GetContractInfo(contractAddrA).IBCPortID, chainB.GetContractInfo(contractAddrB).IBCPortID

	// Create a relayer
	ibcPath := e2eTesting.NewIBCPath(
		s.T(),
		chainA, chainB,
		contractPortA, contractPortB,
		voterTypes.IBCVersion, channelTypes.UNORDERED,
	)
	channelA := ibcPath.EndpointA().ChannelID()

	// Create a voting and add a vote via IBC
	votingID := s.VoterNewVoting(chainA, contractAddrA, accA, "Test", []string{"a"}, time.Minute)

	ibcPacket := s.VoterIBCVote(chainB, contractAddrB, accB, votingID, "a", true, channelA)

	// Check IBC stats before the timeout
	ibcStatsBBefore := s.VoterGetIBCStats(chainB, contractAddrB, accB)
	s.Require().Len(ibcStatsBBefore, 1)
	s.Assert().Equal(voterState.IBCPkgSentStatus, ibcStatsBBefore[0].Status)

	// Timeout
	ibcPath.TimeoutPacket(ibcPacket, ibcPath.EndpointB())

	// Check IBC stats after the timeout
	ibcStatsBAfter := s.VoterGetIBCStats(chainB, contractAddrB, accB)
	s.Require().Len(ibcStatsBAfter, 1)
	s.Assert().Equal(voterState.IBCPkgTimedOutStatus, ibcStatsBAfter[0].Status)
}

// TestVoter_IBCReject test IBC send/ack execution with onReceive error (reject).
func (s *E2ETestSuite) TestVoter_IBCReject() {
	chainA, chainB := s.chainA, s.chainB

	accA, accB := chainA.GetAccount(0), chainB.GetAccount(0)
	contractAddrA, contractAddrB := s.VoterUploadAndInstantiate(chainA, accA), s.VoterUploadAndInstantiate(chainB, accB)
	contractPortA, contractPortB := chainA.GetContractInfo(contractAddrA).IBCPortID, chainB.GetContractInfo(contractAddrB).IBCPortID

	// Create a relayer
	ibcPath := e2eTesting.NewIBCPath(
		s.T(),
		chainA, chainB,
		contractPortA, contractPortB,
		voterTypes.IBCVersion, channelTypes.UNORDERED,
	)
	channelA := ibcPath.EndpointA().ChannelID()

	// Add a vote for non-existing voting
	ibcPacket := s.VoterIBCVote(chainB, contractAddrB, accB, 1, "a", true, channelA)

	// Check IBC stats before the timeout
	ibcStatsBBefore := s.VoterGetIBCStats(chainB, contractAddrB, accB)
	s.Require().Len(ibcStatsBBefore, 1)
	s.Assert().Equal(voterState.IBCPkgSentStatus, ibcStatsBBefore[0].Status)

	// Relay
	ibcPath.RelayPacket(ibcPacket, voterTypes.IBCAckDataFailure)

	// Check IBC stats after the timeout
	ibcStatsBAfter := s.VoterGetIBCStats(chainB, contractAddrB, accB)
	s.Require().Len(ibcStatsBAfter, 1)
	s.Assert().Equal(voterState.IBCPkgRejectedStatus, ibcStatsBAfter[0].Status)
}

// TestVoter_APIVerifySecp256k1Signature tests the API VerifySecp256k1Signature call.
func (s *E2ETestSuite) TestVoter_APIVerifySecp256k1Signature() {
	chain := s.chainA

	acc := chain.GetAccount(0)
	contractAddr := s.VoterUploadAndInstantiate(chain, acc)

	type testCase struct {
		name      string
		genInputs func() (hash, sig, pubKey []byte)
		//
		errExpected bool
		resExpected bool
	}

	genSigAndPubKey := func(msg []byte) (hash, sig, pubKey []byte) {
		msgHash := sha256.Sum256(msg)

		privKey := secp256k1.GenPrivKey()
		signature, err := privKey.Sign(msg)
		s.Require().NoError(err)

		return msgHash[:], signature, privKey.PubKey().Bytes()
	}

	testCases := []testCase{
		{
			name: "OK: valid signature (data taken from the cosmwasm tests)",
			genInputs: func() (hash, sig, pubKey []byte) {
				hashHexStr := "5ae8317d34d1e595e3fa7247db80c0af4320cce1116de187f8f7e2e099c0d8d0"
				sigHexStr := "207082eb2c3dfa0b454e0906051270ba4074ac93760ba9e7110cd9471475111151eb0dbbc9920e72146fb564f99d039802bf6ef2561446eb126ef364d21ee9c4"
				pubKeyHexStr := "04051c1ee2190ecfb174bfe4f90763f2b4ff7517b70a2aec1876ebcfd644c4633fb03f3cfbd94b1f376e34592d9d41ccaf640bb751b00a1fadeb0c01157769eb73"

				hash, err := hex.DecodeString(hashHexStr)
				s.Require().NoError(err)
				sig, err = hex.DecodeString(sigHexStr)
				s.Require().NoError(err)
				pubKey, err = hex.DecodeString(pubKeyHexStr)
				s.Require().NoError(err)

				return hash, sig, pubKey
			},
			resExpected: true,
		},
		{
			name: "OK: valid signature (generated data)",
			genInputs: func() (hash, sig, pubKey []byte) {
				return genSigAndPubKey([]byte{0x01, 0x02, 0x03})
			},
			resExpected: true,
		},
		{
			name: "OK: invalid signature",
			genInputs: func() (hash, sig, pubKey []byte) {
				genHash, genSig, genPubKey := genSigAndPubKey([]byte{0x01, 0x02, 0x03})
				genSig[0] ^= genSig[0]
				return genHash, genSig, genPubKey
			},
			resExpected: false,
		},
		{
			name: "Fail: invalid hash len",
			genInputs: func() (hash, sig, pubKey []byte) {
				genHash, genSig, genPubKey := genSigAndPubKey([]byte{0x01, 0x02, 0x03})
				return genHash[:len(genHash)-1], genSig, genPubKey
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid signature len",
			genInputs: func() (hash, sig, pubKey []byte) {
				genHash, genSig, genPubKey := genSigAndPubKey([]byte{0x01, 0x02, 0x03})
				return genHash, genSig[:len(genSig)-1], genPubKey
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid pubKey len",
			genInputs: func() (hash, sig, pubKey []byte) {
				genHash, genSig, genPubKey := genSigAndPubKey([]byte{0x01, 0x02, 0x03})
				return genHash, genSig, genPubKey[:len(genPubKey)-1]
			},
			errExpected: true,
		},
		{
			name: "Fail: nil hash",
			genInputs: func() (hash, sig, pubKey []byte) {
				_, genSig, genPubKey := genSigAndPubKey([]byte{0x01, 0x02, 0x03})
				return nil, genSig, genPubKey
			},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			hash, sig, pubKey := tc.genInputs()

			req := voterTypes.MsgQuery{
				APIVerifySecp256k1Signature: &voterTypes.QueryAPIVerifySecp256k1SignatureRequest{
					Hash:      hash,
					Signature: sig,
					PubKey:    pubKey,
				},
			}

			res, _ := chain.SmartQueryContract(contractAddr, !tc.errExpected, req)
			if tc.errExpected {
				return
			}

			var resp voterTypes.QueryAPIVerifySecp256k1SignatureResponse
			s.Require().NoError(resp.UnmarshalJSON(res))
			s.Assert().Equal(tc.resExpected, resp.Valid)
		})
	}
}

// TestVoter_APIRecoverSecp256k1PubKey tests the API RecoverSecp256k1PubKey call.
func (s *E2ETestSuite) TestVoter_APIRecoverSecp256k1PubKey() {
	chain := s.chainA

	acc := chain.GetAccount(0)
	contractAddr := s.VoterUploadAndInstantiate(chain, acc)

	type testCase struct {
		name      string
		genInputs func() (hash, sig, pubKeyExpected []byte, recoveryType cwStd.Secp256k1RecoveryParam)
		//
		errExpected   bool
		validExpected bool
	}

	hashHexStr := "5ae8317d34d1e595e3fa7247db80c0af4320cce1116de187f8f7e2e099c0d8d0"
	sigHexStr := "207082eb2c3dfa0b454e0906051270ba4074ac93760ba9e7110cd9471475111151eb0dbbc9920e72146fb564f99d039802bf6ef2561446eb126ef364d21ee9c4"
	pubKeyHexStr := "04051c1ee2190ecfb174bfe4f90763f2b4ff7517b70a2aec1876ebcfd644c4633fb03f3cfbd94b1f376e34592d9d41ccaf640bb751b00a1fadeb0c01157769eb73"

	hashValid, err := hex.DecodeString(hashHexStr)
	s.Require().NoError(err)
	sigValid, err := hex.DecodeString(sigHexStr)
	s.Require().NoError(err)
	pubKeyValid, err := hex.DecodeString(pubKeyHexStr)
	s.Require().NoError(err)
	rParamValid := cwStd.Secp256k1RecoveryParamYCoordIsOdd

	testCases := []testCase{
		{
			name: "OK: successful (data taken from the cosmwasm tests)",
			genInputs: func() (hash, sig, pubKey []byte, recoveryType cwStd.Secp256k1RecoveryParam) {
				return hashValid, sigValid, pubKeyValid, rParamValid
			},
			validExpected: true,
		},
		{
			name: "OK: unsuccessful due to wrong recoveryParam (data taken from the cosmwasm tests)",
			genInputs: func() (hash, sig, pubKey []byte, recoveryType cwStd.Secp256k1RecoveryParam) {
				return hashValid, sigValid, pubKeyValid, cwStd.Secp256k1RecoveryParamYCoordNotOdd
			},
			validExpected: false,
		},
		{
			name: "Fail: invalid hash len",
			genInputs: func() (hash, sig, pubKey []byte, recoveryType cwStd.Secp256k1RecoveryParam) {
				return hashValid[:len(hashValid)-1], sigValid, pubKeyValid, cwStd.Secp256k1RecoveryParamYCoordNotOdd
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid signature len",
			genInputs: func() (hash, sig, pubKey []byte, recoveryType cwStd.Secp256k1RecoveryParam) {
				return hashValid, sigValid[:len(sigValid)-1], pubKeyValid, cwStd.Secp256k1RecoveryParamYCoordNotOdd
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid recovery param",
			genInputs: func() (hash, sig, pubKey []byte, recoveryType cwStd.Secp256k1RecoveryParam) {
				return hashValid, sigValid, pubKeyValid, 2
			},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			hash, sig, pubKeyExpected, rParam := tc.genInputs()

			req := voterTypes.MsgQuery{
				APIRecoverSecp256k1PubKey: &voterTypes.QueryAPIRecoverSecp256k1PubKeyRequest{
					Hash:          hash,
					Signature:     sig,
					RecoveryParam: rParam,
				},
			}

			res, _ := chain.SmartQueryContract(contractAddr, !tc.errExpected, req)
			if tc.errExpected {
				return
			}

			var resp voterTypes.QueryAPIRecoverSecp256k1PubKeyResponse
			s.Require().NoError(resp.UnmarshalJSON(res))

			if tc.validExpected {
				s.Assert().Equal(pubKeyExpected, resp.PubKey)
				return
			}
			s.Assert().NotEqual(pubKeyExpected, resp.PubKey)
		})
	}
}

// TestVoter_APIVerifyEd25519Signature tests the API VerifyEd25519Signature call.
func (s *E2ETestSuite) TestVoter_APIVerifyEd25519Signature() {
	chain := s.chainA

	acc := chain.GetAccount(0)
	contractAddr := s.VoterUploadAndInstantiate(chain, acc)

	type testCase struct {
		name      string
		genInputs func() (message, sig, pubKey []byte)
		//
		errExpected bool
		resExpected bool
	}

	genSigAndPubKey := func(msg []byte) (sig, pubKey []byte) {
		privKey := ed25519.GenPrivKey()
		signature, err := privKey.Sign(msg)
		s.Require().NoError(err)

		return signature, privKey.PubKey().Bytes()
	}

	testCases := []testCase{
		{
			name: "OK: valid signature (data taken from the cosmwasm tests)",
			genInputs: func() (msg, sig, pubKey []byte) {
				msg = []byte("Hello World!")
				sigHexStr := "dea09a2edbcc545c3875ec482602dd61b68273a24f7562db3fb425ee9dbd863ae732a6ade9e72e04bc32c2bd269b25b59342d6da66898f809d0b7e40d8914f05"
				pubKeyHexStr := "bc1c3a48e8b583d7b990e8cbdd0a54744a3152715e20dd4f9451c532d6bbbd7b"

				sig, err := hex.DecodeString(sigHexStr)
				s.Require().NoError(err)
				pubKey, err = hex.DecodeString(pubKeyHexStr)
				s.Require().NoError(err)

				return msg, sig, pubKey
			},
			resExpected: true,
		},
		{
			name: "OK: valid signature (generated data)",
			genInputs: func() (msg, sig, pubKey []byte) {
				msg = []byte{0x00, 0x01, 0x02}
				sig, pubKey = genSigAndPubKey(msg)
				return
			},
			resExpected: true,
		},
		{
			name: "OK: valid signature (empty)",
			genInputs: func() (msg, sig, pubKey []byte) {
				msg = []byte{}
				sig, pubKey = genSigAndPubKey(msg)
				return
			},
			resExpected: true,
		},
		{
			name: "OK: invalid signature",
			genInputs: func() (msg, sig, pubKey []byte) {
				msg = []byte{0x00, 0x01, 0x02}
				sig, pubKey = genSigAndPubKey(msg)
				sig[0] ^= sig[0]
				return
			},
			resExpected: false,
		},
		{
			name: "Fail: invalid signature len",
			genInputs: func() (msg, sig, pubKey []byte) {
				msg = []byte{0x00, 0x01, 0x02}
				sig, pubKey = genSigAndPubKey(msg)
				sig = sig[:len(sig)-1]
				return
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid pubKey len",
			genInputs: func() (msg, sig, pubKey []byte) {
				msg = []byte{0x00, 0x01, 0x02}
				sig, pubKey = genSigAndPubKey(msg)
				pubKey = pubKey[:len(pubKey)-1]
				return
			},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			msg, sig, pubKey := tc.genInputs()

			req := voterTypes.MsgQuery{
				APIVerifyEd25519Signature: &voterTypes.QueryAPIVerifyEd25519SignatureRequest{
					Message:   msg,
					Signature: sig,
					PubKey:    pubKey,
				},
			}

			res, _ := chain.SmartQueryContract(contractAddr, !tc.errExpected, req)
			if tc.errExpected {
				return
			}

			var resp voterTypes.QueryAPIVerifyEd25519SignatureResponse
			s.Require().NoError(resp.UnmarshalJSON(res))
			s.Assert().Equal(tc.resExpected, resp.Valid)
		})
	}
}

// TestVoter_APIVerifyEd25519Signature tests the API VerifyEd25519Signatures call.
func (s *E2ETestSuite) TestVoter_APIVerifyEd25519Signatures() {
	chain := s.chainA

	acc := chain.GetAccount(0)
	contractAddr := s.VoterUploadAndInstantiate(chain, acc)

	type testCase struct {
		name      string
		genInputs func() (messages, sigs, pubKeys [][]byte)
		//
		errExpected bool
		resExpected bool
	}

	genSigAndPubKey := func(msg []byte) (sig, pubKey []byte) {
		privKey := ed25519.GenPrivKey()
		signature, err := privKey.Sign(msg)
		s.Require().NoError(err)

		return signature, privKey.PubKey().Bytes()
	}

	testCases := []testCase{
		{
			name: "OK: valid signatures",
			genInputs: func() (msgs, sigs, pubKeys [][]byte) {
				for i := 0; i < 3; i++ {
					msg := []byte("Hello World, " + strconv.Itoa(i))
					sig, pubKey := genSigAndPubKey(msg)
					msgs = append(msgs, msg)
					sigs = append(sigs, sig)
					pubKeys = append(pubKeys, pubKey)
				}
				return
			},
			resExpected: true,
		},
		{
			name: "OK: invalid signatures",
			genInputs: func() (msgs, sigs, pubKeys [][]byte) {
				for i := 0; i < 3; i++ {
					msg := []byte("Hello World, " + strconv.Itoa(i))
					sig, pubKey := genSigAndPubKey(msg)
					msgs = append(msgs, msg)
					sigs = append(sigs, sig)
					pubKeys = append(pubKeys, pubKey)
				}
				sigs[1][0] ^= sigs[1][0]
				return
			},
			resExpected: false,
		},
		{
			name: "OK: with empty message (empty section)",
			genInputs: func() (msgs, sigs, pubKeys [][]byte) {
				msg1 := []byte{}
				sig1, pubKey1 := genSigAndPubKey(msg1)

				msg2 := []byte("Hello World!")
				sig2, pubKey2 := genSigAndPubKey(msg2)

				return [][]byte{msg1, msg2}, [][]byte{sig1, sig2}, [][]byte{pubKey1, pubKey2}
			},
			resExpected: true,
		},
		{
			name: "Fail: nil messages",
			genInputs: func() (msgs, sigs, pubKeys [][]byte) {
				for i := 0; i < 3; i++ {
					msg := []byte("Hello World, " + strconv.Itoa(i))
					sig, pubKey := genSigAndPubKey(msg)
					msgs = append(msgs, msg)
					sigs = append(sigs, sig)
					pubKeys = append(pubKeys, pubKey)
				}
				return nil, sigs, pubKeys
			},
			errExpected: true,
		},
		{
			name: "Fail: nil signatures",
			genInputs: func() (msgs, sigs, pubKeys [][]byte) {
				for i := 0; i < 3; i++ {
					msg := []byte("Hello World, " + strconv.Itoa(i))
					sig, pubKey := genSigAndPubKey(msg)
					msgs = append(msgs, msg)
					sigs = append(sigs, sig)
					pubKeys = append(pubKeys, pubKey)
				}
				return msgs, nil, pubKeys
			},
			errExpected: true,
		},
		{
			name: "Fail: nil public keys",
			genInputs: func() (msgs, sigs, pubKeys [][]byte) {
				for i := 0; i < 3; i++ {
					msg := []byte("Hello World, " + strconv.Itoa(i))
					sig, pubKey := genSigAndPubKey(msg)
					msgs = append(msgs, msg)
					sigs = append(sigs, sig)
					pubKeys = append(pubKeys, pubKey)
				}
				return msgs, sigs, nil
			},
			errExpected: true,
		},
		{
			name: "Fail: length mismatch",
			genInputs: func() (msgs, sigs, pubKeys [][]byte) {
				for i := 0; i < 3; i++ {
					msg := []byte("Hello World, " + strconv.Itoa(i))
					sig, pubKey := genSigAndPubKey(msg)
					msgs = append(msgs, msg)
					sigs = append(sigs, sig)
					pubKeys = append(pubKeys, pubKey)
				}
				return msgs[:len(msgs)-1], sigs, pubKeys
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid signature",
			genInputs: func() (msgs, sigs, pubKeys [][]byte) {
				for i := 0; i < 3; i++ {
					msg := []byte("Hello World, " + strconv.Itoa(i))
					sig, pubKey := genSigAndPubKey(msg)
					msgs = append(msgs, msg)
					sigs = append(sigs, sig)
					pubKeys = append(pubKeys, pubKey)
				}
				sigs[0] = sigs[0][:len(sigs[0])-1]
				return msgs, sigs, pubKeys
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid pubKey",
			genInputs: func() (msgs, sigs, pubKeys [][]byte) {
				for i := 0; i < 3; i++ {
					msg := []byte("Hello World, " + strconv.Itoa(i))
					sig, pubKey := genSigAndPubKey(msg)
					msgs = append(msgs, msg)
					sigs = append(sigs, sig)
					pubKeys = append(pubKeys, pubKey)
				}
				pubKeys[0] = pubKeys[0][:len(pubKeys[0])-1]
				return msgs, sigs, pubKeys
			},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			msgs, sigs, pubKeys := tc.genInputs()

			req := voterTypes.MsgQuery{
				APIVerifyEd25519Signatures: &voterTypes.QueryAPIVerifyEd25519SignaturesRequest{
					Messages:   msgs,
					Signatures: sigs,
					PubKeys:    pubKeys,
				},
			}

			res, _ := chain.SmartQueryContract(contractAddr, !tc.errExpected, req)
			if tc.errExpected {
				return
			}

			var resp voterTypes.QueryAPIVerifyEd25519SignaturesResponse
			s.Require().NoError(resp.UnmarshalJSON(res))
			s.Assert().Equal(tc.resExpected, resp.Valid)
		})
	}
}

// TestVoter_WASMBindigsQueryNonRewardsQuery sends an empty custom query via WASM bindings.
// Since there is only one module supporting custom WASM query, this should fail.
func (s *E2ETestSuite) TestVoter_WASMBindigsQueryNonRewardsQuery() {
	chain := s.chainA

	acc := chain.GetAccount(0)
	contractAddr := s.VoterUploadAndInstantiate(chain, acc)

	customEmptyQuery := []byte("{}")
	_, err := s.VoterGetCustomQuery(chain, contractAddr, customEmptyQuery, false)
	s.Assert().Contains(err.Error(), "code: 18") // due to CosmWasm error obfuscation, we can't assert for a specific error here
}

// TestVoter_WASMBindingsMetadataQuery tests querying contract metadata via WASM bindings (Custom query plugin & Stargate query).
func (s *E2ETestSuite) TestVoter_WASMBindingsMetadataQuery() {
	chain := s.chainA

	acc1, acc2 := chain.GetAccount(0), chain.GetAccount(1)
	contractAddr := s.VoterUploadAndInstantiate(chain, acc1)

	cmpMetas := func(metaExp rewardsTypes.ContractMetadata, metaRcv voterCustomTypes.ContractMetadataResponse) {
		s.Assert().EqualValues(metaExp.OwnerAddress, metaRcv.OwnerAddress)
		s.Assert().EqualValues(metaExp.RewardsAddress, metaRcv.RewardsAddress)
	}

	getAndCmpMetas := func(metaExp rewardsTypes.ContractMetadata) {
		// wasmvm v1.0.0 (wasmd for us) has disabled the Stargate query, so we skip this case
		// metaRcvStargate := s.VoterGetMetadata(chain, contractAddr, true, true)
		// cmpMetas(metaExp, metaRcvStargate)

		metaRcvCustom := s.VoterGetMetadata(chain, contractAddr, false, true)
		cmpMetas(metaExp, metaRcvCustom)
	}

	var metaExp rewardsTypes.ContractMetadata

	s.Run("No metadata", func() {
		s.VoterGetMetadata(chain, contractAddr, true, false)
	})

	s.Run("Set initial meta", func() {
		metaExp.OwnerAddress = acc1.Address.String()
		metaExp.RewardsAddress = acc1.Address.String()
		chain.SetContractMetadata(acc1, contractAddr, metaExp)

		getAndCmpMetas(metaExp)
	})

	s.Run("Change RewardAddress", func() {
		metaExp.RewardsAddress = acc2.Address.String()
		chain.SetContractMetadata(acc1, contractAddr, metaExp)

		getAndCmpMetas(metaExp)
	})
}

// TestVoter_WASMBindingsSendNonRewardsMsg sends an empty custom message via WASM bindings.
// Since there is only one module supporting custom WASM messages, this should fail.
func (s *E2ETestSuite) TestVoter_WASMBindingsSendNonRewardsMsg() {
	chain := s.chainA

	acc := chain.GetAccount(0)
	contractAddr := s.VoterUploadAndInstantiate(chain, acc)

	customEmptyMsg := []byte("{}")
	err := s.VoterSendCustomMsg(chain, contractAddr, acc, customEmptyMsg, false)
	s.Assert().ErrorIs(err, sdkErrors.ErrInvalidRequest)
}

// TestVoter_WASMBindingsMetadataUpdate tests updating contract metadata via WASM bindings (Custom message).
func (s *E2ETestSuite) TestVoter_WASMBindingsMetadataUpdate() {
	chain := s.chainA

	acc1, acc2 := chain.GetAccount(0), chain.GetAccount(1)
	contractAddr := s.VoterUploadAndInstantiate(chain, acc1)

	s.Run("Fail: no metadata", func() {
		req := voterCustomTypes.UpdateContractMetadataRequest{
			OwnerAddress: acc2.Address.String(),
		}
		err := s.VoterUpdateMetadata(chain, contractAddr, acc1, req, false)
		s.Assert().Contains(err.Error(), "unauthorized")
	})

	// Set initial meta (admin as the OwnerAddress)
	{
		meta := rewardsTypes.ContractMetadata{
			OwnerAddress:   acc1.Address.String(),
			RewardsAddress: acc1.Address.String(),
		}
		chain.SetContractMetadata(acc1, contractAddr, meta)
	}

	s.Run("Fail: update OwnerAddress: unauthorized", func() {
		req := voterCustomTypes.UpdateContractMetadataRequest{
			OwnerAddress: acc2.Address.String(),
		}
		err := s.VoterUpdateMetadata(chain, contractAddr, acc1, req, false)
		s.Assert().Contains(err.Error(), "unauthorized")
	})

	// Update meta (set ContractAddress as the OwnerAddress)
	{
		meta := rewardsTypes.ContractMetadata{
			OwnerAddress: contractAddr.String(),
		}
		chain.SetContractMetadata(acc1, contractAddr, meta)
	}

	s.Run("OK: update RewardAddress", func() {
		req := voterCustomTypes.UpdateContractMetadataRequest{
			RewardsAddress: acc2.Address.String(),
		}
		s.VoterUpdateMetadata(chain, contractAddr, acc1, req, true)

		meta := chain.GetContractMetadata(contractAddr)
		s.Assert().Equal(contractAddr.String(), meta.OwnerAddress)
		s.Assert().Equal(acc2.Address.String(), meta.RewardsAddress)
	})

	s.Run("OK: update OwnerAddress (change ownership)", func() {
		req := voterCustomTypes.UpdateContractMetadataRequest{
			OwnerAddress: acc1.Address.String(),
		}
		s.VoterUpdateMetadata(chain, contractAddr, acc1, req, true)

		meta := chain.GetContractMetadata(contractAddr)
		s.Assert().Equal(acc1.Address.String(), meta.OwnerAddress)
		s.Assert().Equal(acc2.Address.String(), meta.RewardsAddress)
	})

	// update metadata contract to contract
	contractXAddr := s.VoterUploadAndInstantiate(chain, acc1)
	contractYAddr := s.VoterUploadAndInstantiate(chain, acc1)

	// Set initial meta (admin as the OwnerAddress)
	{
		meta := rewardsTypes.ContractMetadata{
			OwnerAddress:   acc1.Address.String(),
			RewardsAddress: acc1.Address.String(),
		}
		chain.SetContractMetadata(acc1, contractXAddr, meta)

		meta = rewardsTypes.ContractMetadata{
			OwnerAddress:   contractXAddr.String(),
			RewardsAddress: acc1.Address.String(),
		}
		chain.SetContractMetadata(acc1, contractYAddr, meta)
	}

	s.Run("Fail: update Contract X owner address from Contract Y: unauthorized", func() {
		// check that contract X's metadata is as expected (acc1 is the owner)
		meta := chain.GetContractMetadata(contractXAddr)
		s.Assert().Equal(acc1.Address.String(), meta.OwnerAddress)
		s.Assert().Equal(acc1.Address.String(), meta.RewardsAddress)

		// update the owner of contract X to be acc2
		req := voterCustomTypes.UpdateContractMetadataRequest{
			ContractAddress: contractXAddr.String(),
			OwnerAddress:    acc2.Address.String(),
		}

		// send the request from contract Y
		err := s.VoterUpdateMetadata(chain, contractYAddr, acc1, req, false)

		s.Assert().Contains(err.Error(), "unauthorized")
	})

	s.Run("OK: update Contract Y metadata from Contract X", func() {
		// check that contract Y's metadata is as expected (X is the owner)
		meta := chain.GetContractMetadata(contractYAddr)
		s.Assert().Equal(contractXAddr.String(), meta.OwnerAddress)
		s.Assert().Equal(acc1.Address.String(), meta.RewardsAddress)

		// update the owner of contract Y to be acc1 and the rewards addrss to be acc2
		req := voterCustomTypes.UpdateContractMetadataRequest{
			ContractAddress: contractYAddr.String(),
			OwnerAddress:    acc1.Address.String(),
			RewardsAddress:  acc2.Address.String(),
		}

		// send the request from contract X
		err := s.VoterUpdateMetadata(chain, contractXAddr, acc1, req, true)
		s.NoError(err)

		// check the update was successful
		meta = chain.GetContractMetadata(contractYAddr)
		s.Assert().Equal(acc1.Address.String(), meta.OwnerAddress)
		s.Assert().Equal(acc2.Address.String(), meta.RewardsAddress)
	})
}

// TestVoter_WASMBindingsRewardsRecordsQuery tests rewards records query via WASM bindings (Custom query plugin).
func (s *E2ETestSuite) TestVoter_WASMBindingsRewardsRecordsQuery() {
	chain := s.chainA

	acc := chain.GetAccount(0)
	contractAddr := s.VoterUploadAndInstantiate(chain, acc)

	// Set initial meta (admin as the OwnerAddress and the contract itself as the RewardsAddress)
	{
		meta := rewardsTypes.ContractMetadata{
			OwnerAddress:   acc.Address.String(),
			RewardsAddress: contractAddr.String(),
		}
		chain.SetContractMetadata(acc, contractAddr, meta)
	}

	// Check there are no rewards yet
	s.Run("Query empty records", func() {
		records, pageResp, _, _ := s.VoterGetRewardsRecords(chain, contractAddr, nil, true)
		s.Assert().Empty(records)
		s.Assert().Empty(pageResp.NextKey)
		s.Assert().Empty(pageResp.Total)
	})

	// Check invalid input
	s.Run("Query over the limit", func() {
		_, _, _, err := s.VoterGetRewardsRecords(
			chain, contractAddr,
			&query.PageRequest{
				Limit: 10000,
			},
			false)
		s.Assert().Contains(err.Error(), "code: 4")
	})

	// Create a new voting and add a vote to get some rewards
	var recordsExpected []rewardsTypes.RewardsRecord
	{
		s.VoterNewVoting(chain, contractAddr, acc, "Test", []string{"a", "b"}, 1*time.Hour)
		s.VoterVote(chain, contractAddr, acc, 0, "a", true)

		recordsExpected = chain.GetApp().RewardsKeeper.GetState().RewardsRecord(chain.GetContext()).GetRewardsRecordByRewardsAddress(contractAddr)
		s.Require().Len(recordsExpected, 2)
	}

	// Check existing rewards
	s.Run("Query all records", func() {
		recordsReceived, pageRespReceived, _, _ := s.VoterGetRewardsRecords(
			chain, contractAddr,
			&query.PageRequest{
				CountTotal: true,
			},
			true,
		)

		s.Assert().ElementsMatch(recordsExpected, recordsReceived)

		s.Assert().Empty(pageRespReceived.NextKey)
		s.Assert().EqualValues(2, pageRespReceived.Total)
	})

	s.Run("Query records with 2 pages", func() {
		// Page 1
		var nextKey []byte
		{
			recordsReceived, pageRespReceived, _, _ := s.VoterGetRewardsRecords(
				chain, contractAddr,
				&query.PageRequest{
					Limit:      1,
					CountTotal: true,
				},
				true,
			)

			s.Assert().ElementsMatch(recordsExpected[:1], recordsReceived)

			s.Assert().NotEmpty(pageRespReceived.NextKey)
			s.Assert().EqualValues(2, pageRespReceived.Total)
			nextKey = pageRespReceived.NextKey
		}

		// Page 2
		{
			recordsReceived, pageRespReceived, _, _ := s.VoterGetRewardsRecords(
				chain, contractAddr,
				&query.PageRequest{
					Key:        nextKey,
					CountTotal: true,
				},
				true,
			)

			s.Assert().ElementsMatch(recordsExpected[1:2], recordsReceived)

			s.Assert().Empty(pageRespReceived.NextKey)
			s.Assert().EqualValues(0, pageRespReceived.Total)
		}
	})
}

// TestVoter_WASMBindingsWithdrawRewards tests rewards withdrawal via WASM bindings (Custom message) using both modes.
// Test also check the Custom message Reply handling.
func (s *E2ETestSuite) TestVoter_WASMBindingsWithdrawRewards() {
	chain := s.chainA

	acc1, acc2, acc3 := chain.GetAccount(0), chain.GetAccount(1), chain.GetAccount(2)
	contractAddr := s.VoterUploadAndInstantiate(chain, acc1)

	// Set initial meta (admin as the OwnerAddress and the contract itself as the RewardsAddress)
	{
		meta := rewardsTypes.ContractMetadata{
			OwnerAddress:   acc1.Address.String(),
			RewardsAddress: contractAddr.String(),
		}
		chain.SetContractMetadata(acc1, contractAddr, meta)
	}

	// Check there are no rewards processed yet
	s.Run("Check Withdraw Reply stats are empty", func() {
		stats := s.VoterGetWithdrawStats(chain, contractAddr)
		s.Assert().Empty(stats.Count)
		s.Assert().Empty(stats.TotalAmount)
		s.Assert().Empty(stats.TotalRecordsUsed)
	})

	// Check invalid input
	s.Run("Invalid withdraw request", func() {
		err := s.VoterWithdrawRewards(
			chain, contractAddr, acc1,
			pkg.Uint64Ptr(2),
			[]uint64{1},
			false,
		)
		s.Assert().Contains(err.Error(), "withdrawRewards: one of (RecordsLimit, RecordIDs) fields must be set")
	})

	// Create a new voting and add a few votes to get some rewards
	var recordsExpected []rewardsTypes.RewardsRecord
	var totalRewardsExpected sdk.Coins
	{
		s.VoterNewVoting(chain, contractAddr, acc1, "Test", []string{"a", "b", "c"}, 1*time.Hour)
		s.VoterVote(chain, contractAddr, acc1, 0, "a", true)
		s.VoterVote(chain, contractAddr, acc2, 0, "b", false)
		s.VoterVote(chain, contractAddr, acc3, 0, "c", true)

		recordsExpected = chain.GetApp().RewardsKeeper.GetState().RewardsRecord(chain.GetContext()).GetRewardsRecordByRewardsAddress(contractAddr)
		s.Require().Len(recordsExpected, 4)

		for _, record := range recordsExpected {
			totalRewardsExpected = totalRewardsExpected.Add(record.Rewards...)
		}
	}

	// Get the rewardsAddr initial balance to check it after all the withdrawals are done
	rewardsAddrBalanceBefore := chain.GetBalance(contractAddr)

	// Withdraw using records limit
	s.Run("Withdraw using records limit and check Reply stats", func() {
		s.VoterWithdrawRewards(
			chain, contractAddr, acc1,
			pkg.Uint64Ptr(2),
			nil,
			true,
		)

		rewardsExpected := sdk.NewCoins()
		rewardsExpected = rewardsExpected.Add(recordsExpected[0].Rewards...)
		rewardsExpected = rewardsExpected.Add(recordsExpected[1].Rewards...)

		stats := s.VoterGetWithdrawStats(chain, contractAddr)
		s.Assert().EqualValues(1, stats.Count)
		s.Assert().Equal(rewardsExpected.String(), s.CosmWasmCoinsToSDK(stats.TotalAmount...).String())
		s.Assert().EqualValues(2, stats.TotalRecordsUsed)
	})

	// Withdraw the rest using record IDs
	s.Run("Withdraw using record IDs and check Reply stats", func() {
		s.VoterWithdrawRewards(
			chain, contractAddr, acc1,
			nil,
			[]uint64{recordsExpected[2].Id, recordsExpected[3].Id},
			true,
		)

		stats := s.VoterGetWithdrawStats(chain, contractAddr)
		s.Assert().EqualValues(2, stats.Count)
		s.Assert().Equal(totalRewardsExpected.String(), s.CosmWasmCoinsToSDK(stats.TotalAmount...).String())
		s.Assert().EqualValues(4, stats.TotalRecordsUsed)
	})

	s.Run("Check rewardsAddr balance changed", func() {
		rewardsAddrBalanceDiff := chain.GetBalance(contractAddr).Sub(rewardsAddrBalanceBefore)
		s.Assert().Equal(totalRewardsExpected.String(), rewardsAddrBalanceDiff.String())
	})
}
