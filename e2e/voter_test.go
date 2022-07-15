package e2e

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"

	voterPkg "github.com/CosmWasm/cosmwasm-go/example/voter/src/pkg"
	voterState "github.com/CosmWasm/cosmwasm-go/example/voter/src/state"
	voterTypes "github.com/CosmWasm/cosmwasm-go/example/voter/src/types"
	cwStd "github.com/CosmWasm/cosmwasm-go/std"
	cwTypes "github.com/CosmWasm/cosmwasm-go/std/types"
	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channelTypes "github.com/cosmos/ibc-go/v2/modules/core/04-channel/types"
)

const (
	VoterWasmPath       = "contracts/voter.wasm"
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

		releasedCoinsRcv := s.VoterRelease(chain, contractAddr, acc1)
		s.Assert().EqualValues(contractCoinsExp, releasedCoinsRcv)

		acc1BalanceAfter := chain.GetBalance(acc1.Address)
		s.Assert().EqualValues(acc1BalanceBefore.Add(contractCoinsExp...), acc1BalanceAfter)

		releaseStats := s.VoterGetReleaseStats(chain, contractAddr)
		s.Assert().EqualValues(1, releaseStats.Count)
		s.Assert().EqualValues(releasedCoinsRcv, s.CosmWasmCoinsToSDK(releaseStats.TotalAmount...))
	})
}

// TestVoter_Sudo tests Sudo execution via Gov proposal to change the contract parameters.
// Test indirectly checks:
//   - api.HumanAddress: called by querying contract Params (OwnerAddr);
//   - and api.CanonicalAddress: called by instantiate to convert OwnerAddr (string) to canonical address (bytes);
func (s *E2ETestSuite) TestVoter_Sudo() {
	// x/wasmd/types/codec.go:
	//   registry.RegisterImplementations() call:
	//     "&SudoContractProposal{}" should be added
	s.T().Skip("Current wasmd dependency doesn't have SudoContractProposal type registered (fixed in 1.0.0, skip for now)")

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
