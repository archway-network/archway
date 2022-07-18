// DONTCOVER
package e2eTesting

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/client"
	"strconv"
	"testing"
	"time"

	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/archway-network/archway/app"
	"github.com/cosmos/cosmos-sdk/codec"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptoCodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptoTypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/ibc-go/v2/testing/mock"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmProto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmTypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

// TestChain keeps a test chain state and provides helper functions to simulate various operations.
// Heavily inspired by the TestChain from the ibc-go repo (https://github.com/cosmos/ibc-go/blob/main/testing/chain.go).
// Reasons for creating a custom TestChain rather than using the ibc-go's one are: to simplify it,
// add contract related helpers and fix errors caused by x/gastracker module (ibc-go version starts at block 2).
type TestChain struct {
	t *testing.T

	app         *app.ArchwayApp         // main application
	lastHeader  tmProto.Header          // header for the last committed block
	curHeader   tmProto.Header          // header for the current block
	txConfig    client.TxConfig         // config to sing TXs
	valSet      *tmTypes.ValidatorSet   // validator set for the current block
	valSigners  []tmTypes.PrivValidator // validator signers for the current block
	accPrivKeys []cryptoTypes.PrivKey   // genesis account private keys
}

// NewTestChain creates a new TestChain with the default amount of genesis accounts and validators.
func NewTestChain(t *testing.T, chainIdx int) *TestChain {
	const (
		validatorsN      = 1
		genAccsN         = 5
		genBalanceAmount = "1000000000"
		bondAmount       = "1000000"
		chainIDPrefix    = "test-"
	)

	// Create an app and a default genesis state
	encCfg := app.MakeEncodingConfig()

	//logger := log.TestingLogger()
	logger := log.TestingLogger()

	archApp := app.NewArchwayApp(
		logger,
		dbm.NewMemDB(),
		nil,
		true, map[int64]bool{},
		app.DefaultNodeHome,
		5,
		encCfg,
		app.GetEnabledProposals(),
		app.EmptyBaseAppOptions{},
		[]wasm.Option{},
	)
	genState := app.NewDefaultGenesisState()

	// Generate validators
	validators := make([]*tmTypes.Validator, 0, validatorsN)
	valSigners := make([]tmTypes.PrivValidator, 0, validatorsN)
	for i := 0; i < validatorsN; i++ {
		valPrivKey := mock.NewPV()
		valPubKey, err := valPrivKey.GetPubKey()
		require.NoError(t, err)

		validators = append(validators, tmTypes.NewValidator(valPubKey, 1))
		valSigners = append(valSigners, valPrivKey)
	}
	validatorSet := tmTypes.NewValidatorSet(validators)

	// Generate genesis accounts, gen and bond coins
	genAccs := make([]authTypes.GenesisAccount, 0, genAccsN)
	genAccPrivKeys := make([]cryptoTypes.PrivKey, 0, genAccsN)
	for i := 0; i < genAccsN; i++ {
		accPrivKey := secp256k1.GenPrivKey()
		acc := authTypes.NewBaseAccount(accPrivKey.PubKey().Address().Bytes(), accPrivKey.PubKey(), uint64(i), 0)

		genAccs = append(genAccs, acc)
		genAccPrivKeys = append(genAccPrivKeys, accPrivKey)
	}

	genAmt, ok := sdk.NewIntFromString(genBalanceAmount)
	require.True(t, ok)
	genCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, genAmt))

	bondAmt, ok := sdk.NewIntFromString(bondAmount)
	require.True(t, ok)
	bondCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, bondAmt))

	// Update the x/auth genesis with gen accounts
	authGenesis := authTypes.NewGenesisState(authTypes.DefaultParams(), genAccs)
	genState[authTypes.ModuleName] = archApp.AppCodec().MustMarshalJSON(authGenesis)

	// Update the x/staking genesis (every gen account is a corresponding validator's delegator)
	stakingValidators := make([]stakingTypes.Validator, 0, len(validatorSet.Validators))
	stakingDelegations := make([]stakingTypes.Delegation, 0, len(validatorSet.Validators))
	for i, val := range validatorSet.Validators {
		valPubKey, err := cryptoCodec.FromTmPubKeyInterface(val.PubKey)
		require.NoError(t, err)

		valPubKeyAny, err := codecTypes.NewAnyWithValue(valPubKey)
		require.NoError(t, err)

		validator := stakingTypes.Validator{
			OperatorAddress:   sdk.ValAddress(val.Address).String(),
			ConsensusPubkey:   valPubKeyAny,
			Jailed:            false,
			Status:            stakingTypes.Bonded,
			Tokens:            bondAmt,
			DelegatorShares:   sdk.OneDec(),
			Description:       stakingTypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingTypes.NewCommission(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
			MinSelfDelegation: sdk.ZeroInt(),
		}

		stakingValidators = append(stakingValidators, validator)
		stakingDelegations = append(stakingDelegations, stakingTypes.NewDelegation(genAccs[i].GetAddress(), val.Address.Bytes(), sdk.OneDec()))
	}

	stakingGenesis := stakingTypes.NewGenesisState(stakingTypes.DefaultParams(), stakingValidators, stakingDelegations)
	genState[stakingTypes.ModuleName] = archApp.AppCodec().MustMarshalJSON(stakingGenesis)

	// Update x/bank genesis with total supply, gen account balances and bonding pool balance
	totalSupply := sdk.NewCoins()
	bondedPoolCoins := sdk.NewCoins()
	balances := make([]bankTypes.Balance, 0, genAccsN)
	for i := 0; i < genAccsN; i++ {
		balances = append(balances, bankTypes.Balance{
			Address: genAccs[i].GetAddress().String(),
			Coins:   genCoins,
		})
		totalSupply = totalSupply.Add(genCoins...)
	}
	for i := 0; i < validatorsN; i++ {
		bondedPoolCoins = bondedPoolCoins.Add(bondCoins...)
		totalSupply = totalSupply.Add(bondCoins...)
	}
	balances = append(balances, bankTypes.Balance{
		Address: authTypes.NewModuleAddress(stakingTypes.BondedPoolName).String(),
		Coins:   bondedPoolCoins,
	})

	bankGenesis := bankTypes.NewGenesisState(bankTypes.DefaultGenesisState().Params, balances, totalSupply, []bankTypes.Metadata{})
	genState[bankTypes.ModuleName] = archApp.AppCodec().MustMarshalJSON(bankGenesis)

	// Init chain
	genStateBytes, err := json.MarshalIndent(genState, "", " ")
	require.NoError(t, err)

	archApp.InitChain(
		abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: app.DefaultConsensusParams,
			AppStateBytes:   genStateBytes,
		},
	)

	// Create a chain and finalize the 1st block
	chain := TestChain{
		t:   t,
		app: archApp,
		curHeader: tmProto.Header{
			ChainID: chainIDPrefix + strconv.Itoa(chainIdx),
			Time:    time.Unix(0, 0).UTC(),
		},
		txConfig:    encCfg.TxConfig,
		valSet:      validatorSet,
		valSigners:  valSigners,
		accPrivKeys: genAccPrivKeys,
	}
	chain.beginBlock()
	chain.endBlock()

	// Start a new block
	chain.beginBlock()
	return &chain
}

// GetAccount returns account address and private key with the given index.
func (chain *TestChain) GetAccount(idx int) Account {
	t := chain.t

	require.Less(t, idx, len(chain.accPrivKeys))
	privKey := chain.accPrivKeys[idx]

	return Account{
		Address: sdk.AccAddress(privKey.PubKey().Address().Bytes()),
		PrivKey: privKey,
	}
}

// GetBalance returns the balance of the given account.
func (chain *TestChain) GetBalance(accAddr sdk.AccAddress) sdk.Coins {
	return chain.app.BankKeeper.GetAllBalances(chain.GetContext(), accAddr)
}

// GetContext returns a context for the current block.
func (chain *TestChain) GetContext() sdk.Context {
	return chain.app.BaseApp.NewContext(false, chain.curHeader)
}

// GetAppCodec returns the application codec.
func (chain *TestChain) GetAppCodec() codec.Codec {
	return chain.app.AppCodec()
}

// GetChainID returns the chain ID.
func (chain *TestChain) GetChainID() string {
	return chain.curHeader.ChainID
}

// GetBlockTime returns the current block time.
func (chain *TestChain) GetBlockTime() time.Time {
	return chain.curHeader.Time
}

// GetBlockHeight returns the current block height.
func (chain *TestChain) GetBlockHeight() int64 {
	return chain.app.LastBlockHeight()
}

// GetUnbondingTime returns x/staking validator unbonding time.
func (chain *TestChain) GetUnbondingTime() time.Duration {
	return chain.app.StakingKeeper.UnbondingTime(chain.GetContext())
}

// NextBlock starts a new block with options time shift.
func (chain *TestChain) NextBlock(skipTime time.Duration) {
	chain.endBlock()

	chain.curHeader.Time = chain.curHeader.Time.Add(skipTime)
	chain.beginBlock()
}

type SendMsgOption func(opt *sendMsgOptions)

type sendMsgOptions struct {
	fees     sdk.Coins
	gasLimit uint64
}

func SendMsgWithFees(coins sdk.Coins) SendMsgOption {
	return func(opt *sendMsgOptions) {
		opt.fees = coins
	}
}

// SendMsgs sends a series of messages.
func (chain *TestChain) SendMsgs(senderAcc Account, expPass bool, msgs []sdk.Msg, opts ...SendMsgOption) (sdk.GasInfo, *sdk.Result, error) {
	options := &sendMsgOptions{
		fees:     sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)),
		gasLimit: 10_000_000,
	}

	for _, o := range opts {
		o(options)
	}

	t := chain.t

	// Get the sender account
	senderAccI := chain.app.AccountKeeper.GetAccount(chain.GetContext(), senderAcc.Address)
	require.NotNil(t, senderAccI)

	// Build and sign Tx
	tx, err := helpers.GenTx(
		chain.txConfig,
		msgs,
		options.fees,
		options.gasLimit,
		chain.GetChainID(),
		[]uint64{senderAccI.GetAccountNumber()},
		[]uint64{senderAccI.GetSequence()},
		senderAcc.PrivKey,
	)
	require.NoError(t, err)

	// Send the Tx
	gasInfo, res, err := chain.app.Deliver(chain.txConfig.TxEncoder(), tx)
	if expPass {
		require.NoError(t, err)
		require.NotNil(t, res)
	} else {
		require.Error(t, err)
		require.Nil(t, res)
	}

	chain.endBlock()
	chain.beginBlock()

	return gasInfo, res, nil
}

// ParseSDKResultData converts TX result data into a slice of Msgs.
func (chain *TestChain) ParseSDKResultData(r *sdk.Result) sdk.TxMsgData {
	t := chain.t

	require.NotNil(t, r)

	var protoResult sdk.TxMsgData
	require.NoError(chain.t, proto.Unmarshal(r.Data, &protoResult))

	return protoResult
}

// beginBlock begins a new block.
func (chain *TestChain) beginBlock() {
	const blockDur = 5 * time.Second

	chain.lastHeader = chain.curHeader

	chain.curHeader.Height++
	chain.curHeader.Time = chain.curHeader.Time.Add(blockDur)
	chain.curHeader.AppHash = chain.app.LastCommitID().Hash
	chain.curHeader.ValidatorsHash = chain.valSet.Hash()
	chain.curHeader.NextValidatorsHash = chain.valSet.Hash()

	chain.app.BeginBlock(abci.RequestBeginBlock{Header: chain.curHeader})
}

// endBlock finalizes the current block.
func (chain *TestChain) endBlock() {
	chain.app.EndBlock(abci.RequestEndBlock{Height: chain.curHeader.Height})
	chain.app.Commit()
}
