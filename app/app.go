package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/archway-network/archway/app/keepers"
	"github.com/archway-network/archway/x/genmsg"
	"github.com/cosmos/cosmos-sdk/server"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	wasmdKeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	cosmwasm "github.com/CosmWasm/wasmvm"
	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/libs/log"
	tmos "github.com/cometbft/cometbft/libs/os"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/posthandler"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govV1Beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"
	"github.com/cosmos/cosmos-sdk/x/mint"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/nft"
	nftkeeper "github.com/cosmos/cosmos-sdk/x/nft/keeper"
	nftmodule "github.com/cosmos/cosmos-sdk/x/nft/module"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ica "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts"
	icahost "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	ibcfee "github.com/cosmos/ibc-go/v7/modules/apps/29-fee"
	ibcfeekeeper "github.com/cosmos/ibc-go/v7/modules/apps/29-fee/keeper"
	ibcfeetypes "github.com/cosmos/ibc-go/v7/modules/apps/29-fee/types"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v7/modules/core"
	ibcclient "github.com/cosmos/ibc-go/v7/modules/core/02-client"
	ibcclientclient "github.com/cosmos/ibc-go/v7/modules/core/02-client/client"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	"github.com/spf13/cast"

	"github.com/archway-network/archway/wasmbinding"
	"github.com/archway-network/archway/x/rewards"
	rewardsKeeper "github.com/archway-network/archway/x/rewards/keeper"
	"github.com/archway-network/archway/x/rewards/mintbankkeeper"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
	"github.com/archway-network/archway/x/tracking"
	trackingKeeper "github.com/archway-network/archway/x/tracking/keeper"
	trackingTypes "github.com/archway-network/archway/x/tracking/types"

	"github.com/CosmWasm/wasmd/x/wasm"

	archwayappparams "github.com/archway-network/archway/app/params"
	archway "github.com/archway-network/archway/types"

	// unnamed import of statik for swagger UI support
	_ "github.com/cosmos/cosmos-sdk/client/docs/statik"
)

const appName = "Archway"

// We pull these out so we can set them with LDFLAGS in the Makefile
var (
	NodeDir      = ".archway"
	Bech32Prefix = "archway"
)

// These constants are derived from the above variables.
// These are the ones we will want to use in the code, based on
// any overrides above
var (
	// DefaultNodeHome default home directories for archwayd
	DefaultNodeHome = os.ExpandEnv("$HOME/") + NodeDir

	// Bech32PrefixAccAddr defines the Bech32 prefix of an account's address
	Bech32PrefixAccAddr = Bech32Prefix
	// Bech32PrefixAccPub defines the Bech32 prefix of an account's public key
	Bech32PrefixAccPub = Bech32Prefix + sdk.PrefixPublic
	// Bech32PrefixValAddr defines the Bech32 prefix of a validator's operator address
	Bech32PrefixValAddr = Bech32Prefix + sdk.PrefixValidator + sdk.PrefixOperator
	// Bech32PrefixValPub defines the Bech32 prefix of a validator's operator public key
	Bech32PrefixValPub = Bech32Prefix + sdk.PrefixValidator + sdk.PrefixOperator + sdk.PrefixPublic
	// Bech32PrefixConsAddr defines the Bech32 prefix of a consensus node address
	Bech32PrefixConsAddr = Bech32Prefix + sdk.PrefixValidator + sdk.PrefixConsensus
	// Bech32PrefixConsPub defines the Bech32 prefix of a consensus node public key
	Bech32PrefixConsPub = Bech32Prefix + sdk.PrefixValidator + sdk.PrefixConsensus + sdk.PrefixPublic
)

var (
	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(GetGovProposalHandlers()),
		groupmodule.AppModuleBasic{},
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		feegrantmodule.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		consensus.AppModuleBasic{},
		ibc.AppModuleBasic{},
		ibctm.AppModuleBasic{},
		ibcfee.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		transfer.AppModuleBasic{},
		vesting.AppModuleBasic{},
		nftmodule.AppModuleBasic{},
		wasm.AppModuleBasic{},
		ica.AppModuleBasic{},
		tracking.AppModuleBasic{},
		rewards.AppModuleBasic{},
		genmsg.AppModule{},
	)

	// module account permissions
	maccPerms = map[string][]string{
		rewardsTypes.ContractRewardCollector: nil,
		authtypes.FeeCollectorName:           {authtypes.Burner},
		distrtypes.ModuleName:                nil,
		minttypes.ModuleName:                 {authtypes.Minter},
		stakingtypes.BondedPoolName:          {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName:       {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:                  {authtypes.Burner},
		nft.ModuleName:                       nil,
		ibctransfertypes.ModuleName:          {authtypes.Minter, authtypes.Burner},
		ibcfeetypes.ModuleName:               nil,
		icatypes.ModuleName:                  nil,
		wasmdTypes.ModuleName:                {authtypes.Burner},
		rewardsTypes.TreasuryCollector:       {authtypes.Burner},
	}
)

var (
	_ runtime.AppI            = (*ArchwayApp)(nil)
	_ servertypes.Application = (*ArchwayApp)(nil)
)

func init() {
	// sets the default power reduction in order to ensure that on high precision numbers, which is a default for archway
	// the network does not get stalled due to an integer overflow in some edge cases.
	sdk.DefaultPowerReduction = archway.DefaultPowerReduction
}

// ArchwayApp extended ABCI application
type ArchwayApp struct {
	*baseapp.BaseApp
	legacyAmino       *codec.LegacyAmino //nolint:staticcheck
	appCodec          codec.Codec
	interfaceRegistry types.InterfaceRegistry

	invCheckPeriod uint

	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// Keepers
	Keepers keepers.ArchwayKeepers

	ScopedIBCKeeper      capabilitykeeper.ScopedKeeper
	ScopedICAHostKeeper  capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper capabilitykeeper.ScopedKeeper
	ScopedIBCFeeKeeper   capabilitykeeper.ScopedKeeper
	ScopedWASMKeeper     capabilitykeeper.ScopedKeeper

	// the module manager
	mm *module.Manager

	// simulation manager
	sm *module.SimulationManager

	// module configurator
	configurator module.Configurator
}

// NewArchwayApp returns a reference to an initialized ArchwayApp.
func NewArchwayApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	encodingConfig archwayappparams.EncodingConfig,
	appOpts servertypes.AppOptions,
	wasmOpts []wasmdKeeper.Option,
	baseAppOptions ...func(*baseapp.BaseApp),
) *ArchwayApp {
	appCodec, legacyAmino := encodingConfig.Marshaler, encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry

	bApp := baseapp.NewBaseApp(appName, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	govModuleAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	keys := sdk.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, ibcexported.StoreKey, upgradetypes.StoreKey,
		evidencetypes.StoreKey, ibctransfertypes.StoreKey, capabilitytypes.StoreKey,
		feegrant.StoreKey, authzkeeper.StoreKey, wasmdTypes.StoreKey, consensusparamtypes.StoreKey,
		icahosttypes.StoreKey, ibcfeetypes.StoreKey, crisistypes.StoreKey, group.StoreKey, nftkeeper.StoreKey,

		trackingTypes.StoreKey, rewardsTypes.StoreKey,
	)
	tkeys := sdk.NewTransientStoreKeys(paramstypes.TStoreKey)
	memKeys := sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	app := &ArchwayApp{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
		keys:              keys,
		tkeys:             tkeys,
		memKeys:           memKeys,
		Keepers:           keepers.ArchwayKeepers{},
	}

	app.Keepers.ParamsKeeper = initParamsKeeper(
		appCodec,
		legacyAmino,
		keys[paramstypes.StoreKey],
		tkeys[paramstypes.TStoreKey],
	)

	// set the BaseApp's parameter store
	app.Keepers.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(appCodec, keys[consensusparamtypes.StoreKey], authtypes.NewModuleAddress(govtypes.ModuleName).String())
	app.SetParamStore(&app.Keepers.ConsensusParamsKeeper)

	// add capability keeper and ScopeToModule for ibc module
	app.Keepers.CapabilityKeeper = capabilitykeeper.NewKeeper(
		appCodec,
		keys[capabilitytypes.StoreKey],
		memKeys[capabilitytypes.MemStoreKey],
	)

	scopedIBCKeeper := app.Keepers.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	scopedICAHostKeeper := app.Keepers.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	scopedTransferKeeper := app.Keepers.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	scopedWasmKeeper := app.Keepers.CapabilityKeeper.ScopeToModule(wasmdTypes.ModuleName)
	app.Keepers.CapabilityKeeper.Seal()

	// add keepers
	app.Keepers.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		keys[authtypes.StoreKey],
		authtypes.ProtoBaseAccount,
		maccPerms,
		Bech32Prefix,
		govModuleAddr,
	)
	app.Keepers.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		keys[banktypes.StoreKey],
		app.Keepers.AccountKeeper,
		BlockedAddresses(),
		govModuleAddr,
	)
	app.Keepers.AuthzKeeper = authzkeeper.NewKeeper(
		keys[authzkeeper.StoreKey],
		appCodec,
		app.BaseApp.MsgServiceRouter(),
		app.Keepers.AccountKeeper,
	)
	app.Keepers.FeeGrantKeeper = feegrantkeeper.NewKeeper(
		appCodec,
		keys[feegrant.StoreKey],
		app.Keepers.AccountKeeper,
	)
	app.Keepers.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec,
		keys[stakingtypes.StoreKey],
		app.Keepers.AccountKeeper,
		app.Keepers.BankKeeper,
		govModuleAddr,
	)
	app.Keepers.DistrKeeper = distrkeeper.NewKeeper(
		appCodec,
		keys[distrtypes.StoreKey],
		app.Keepers.AccountKeeper,
		app.Keepers.BankKeeper,
		app.Keepers.StakingKeeper,
		authtypes.FeeCollectorName,
		govModuleAddr,
	)
	app.Keepers.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec,
		legacyAmino,
		keys[slashingtypes.StoreKey],
		app.Keepers.StakingKeeper,
		govModuleAddr,
	)
	app.Keepers.CrisisKeeper = *crisiskeeper.NewKeeper(
		appCodec,
		keys[crisistypes.StoreKey],
		invCheckPeriod,
		app.Keepers.BankKeeper,
		authtypes.FeeCollectorName,
		govModuleAddr,
	)
	app.Keepers.UpgradeKeeper = *upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		keys[upgradetypes.StoreKey],
		appCodec,
		homePath,
		app.BaseApp,
		govModuleAddr,
	)

	groupConfig := group.DefaultConfig()
	app.Keepers.GroupKeeper = groupkeeper.NewKeeper(keys[group.StoreKey], appCodec, app.MsgServiceRouter(), app.Keepers.AccountKeeper, groupConfig)

	app.Keepers.NFTKeeper = nftkeeper.NewKeeper(
		keys[nftkeeper.StoreKey],
		appCodec,
		app.Keepers.AccountKeeper,
		app.Keepers.BankKeeper,
	)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.Keepers.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(app.Keepers.DistrKeeper.Hooks(), app.Keepers.SlashingKeeper.Hooks()),
	)

	app.Keepers.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		keys[ibcexported.StoreKey],
		app.getSubspace(ibcexported.ModuleName),
		app.Keepers.StakingKeeper,
		app.Keepers.UpgradeKeeper,
		scopedIBCKeeper,
	)

	// register the proposal types
	govRouter := govV1Beta1types.NewRouter()
	govRouter.
		AddRoute(govtypes.RouterKey, govV1Beta1types.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.Keepers.ParamsKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(&app.Keepers.UpgradeKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(app.Keepers.IBCKeeper.ClientKeeper))

	// IBC Fee Module keeper
	app.Keepers.IBCFeeKeeper = ibcfeekeeper.NewKeeper(
		appCodec, keys[ibcfeetypes.StoreKey],
		app.Keepers.IBCKeeper.ChannelKeeper, // may be replaced with IBC middleware
		app.Keepers.IBCKeeper.ChannelKeeper,
		&app.Keepers.IBCKeeper.PortKeeper, app.Keepers.AccountKeeper, app.Keepers.BankKeeper,
	)

	// Create Transfer Keepers
	app.Keepers.TransferKeeper = ibctransferkeeper.NewKeeper(
		appCodec,
		keys[ibctransfertypes.StoreKey],
		app.getSubspace(ibctransfertypes.ModuleName),
		app.Keepers.IBCFeeKeeper, // ISC4 Wrapper: fee IBC middleware
		app.Keepers.IBCKeeper.ChannelKeeper,
		&app.Keepers.IBCKeeper.PortKeeper,
		app.Keepers.AccountKeeper,
		app.Keepers.BankKeeper,
		scopedTransferKeeper)

	transferModule := transfer.NewAppModule(app.Keepers.TransferKeeper)

	app.Keepers.ICAHostKeeper = icahostkeeper.NewKeeper(
		appCodec,
		keys[icahosttypes.StoreKey],
		app.getSubspace(icahosttypes.SubModuleName),
		app.Keepers.IBCFeeKeeper,
		app.Keepers.IBCKeeper.ChannelKeeper,
		&app.Keepers.IBCKeeper.PortKeeper,
		app.Keepers.AccountKeeper,
		scopedICAHostKeeper,
		app.MsgServiceRouter(),
	)

	// create evidence keeper with router
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec,
		keys[evidencetypes.StoreKey],
		app.Keepers.StakingKeeper,
		app.Keepers.SlashingKeeper,
	)
	app.Keepers.EvidenceKeeper = *evidenceKeeper

	defaultGasRegister := wasmdTypes.NewDefaultWasmGasRegister()

	app.Keepers.TrackingKeeper = trackingKeeper.NewKeeper(
		appCodec,
		keys[trackingTypes.StoreKey],
		defaultGasRegister,
	)

	wasmDir := filepath.Join(homePath, "wasm")
	wasmConfig, err := wasm.ReadWasmConfig(appOpts)
	if err != nil {
		panic(fmt.Sprintf("error while reading wasm config: %s", err))
	}

	// The last arguments can contain custom message handlers, and custom query handlers,
	// if we want to allow any custom callbacks
	supportedFeatures := "iterator,staking,stargate,cosmwasm_1_1,cosmwasm_1_2,cosmwasm_1_3,cosmwasm_1_4"

	wasmer, err := cosmwasm.NewVM(filepath.Join(wasmDir, "wasm"), supportedFeatures, 32, wasmConfig.ContractDebugMode, wasmConfig.MemoryCacheSize)
	if err != nil {
		panic(err)
	}

	trackingWasmVm := wasmdTypes.NewTrackingWasmerEngine(wasmer, &wasmdTypes.NoOpContractGasProcessor{})

	wasmOpts = append(wasmOpts, wasmdKeeper.WithWasmEngine(trackingWasmVm), wasmdKeeper.WithGasRegister(defaultGasRegister))
	// Archway specific options (using a pointer as the keeper is post-initialized below)
	wasmOpts = append(wasmOpts, wasmbinding.BuildWasmOptions(&app.Keepers.RewardsKeeper, &app.Keepers.GovKeeper)...)

	app.Keepers.WASMKeeper = wasmdKeeper.NewKeeper(
		appCodec,
		keys[wasmdTypes.StoreKey],
		app.Keepers.AccountKeeper,
		app.Keepers.BankKeeper,
		app.Keepers.StakingKeeper,
		distrkeeper.NewQuerier(app.Keepers.DistrKeeper),
		app.Keepers.IBCFeeKeeper, // ISC4 Wrapper: fee IBC middleware
		app.Keepers.IBCKeeper.ChannelKeeper,
		&app.Keepers.IBCKeeper.PortKeeper,
		scopedWasmKeeper,
		app.Keepers.TransferKeeper,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		wasmDir,
		wasmConfig,
		supportedFeatures,
		govModuleAddr,
		wasmOpts...,
	)

	// Setting gas recorder here to avoid cyclic loop
	trackingWasmVm.SetGasRecorder(app.Keepers.TrackingKeeper)

	app.Keepers.RewardsKeeper = rewardsKeeper.NewKeeper(
		appCodec,
		keys[rewardsTypes.StoreKey],
		app.Keepers.WASMKeeper,
		app.Keepers.TrackingKeeper,
		app.Keepers.AccountKeeper,
		app.Keepers.BankKeeper,
		app.getSubspace(rewardsTypes.ModuleName),
		govModuleAddr,
	)

	// Note we set up mint keeper after the x/rewards keeper
	app.Keepers.MintKeeper = mintkeeper.NewKeeper(
		appCodec,
		keys[minttypes.StoreKey],
		app.Keepers.StakingKeeper,
		app.Keepers.AccountKeeper,
		mintbankkeeper.NewKeeper(app.Keepers.BankKeeper, app.Keepers.RewardsKeeper),
		authtypes.FeeCollectorName,
		govModuleAddr,
	)

	var transferStack porttypes.IBCModule
	transferStack = transfer.NewIBCModule(app.Keepers.TransferKeeper)
	transferStack = ibcfee.NewIBCMiddleware(transferStack, app.Keepers.IBCFeeKeeper)

	// Create Interchain Accounts Stack
	// RecvPacket, message that originates from core IBC and goes down to app, the flow is:
	// channel.RecvPacket -> fee.OnRecvPacket -> icaHost.OnRecvPacket
	var icaHostStack porttypes.IBCModule
	icaHostStack = icahost.NewIBCModule(app.Keepers.ICAHostKeeper)
	icaHostStack = ibcfee.NewIBCMiddleware(icaHostStack, app.Keepers.IBCFeeKeeper)

	var wasmStack porttypes.IBCModule
	wasmStack = wasm.NewIBCHandler(app.Keepers.WASMKeeper, app.Keepers.IBCKeeper.ChannelKeeper, app.Keepers.IBCFeeKeeper)
	wasmStack = ibcfee.NewIBCMiddleware(wasmStack, app.Keepers.IBCFeeKeeper)

	// create static IBC router, add transfer route, add wasm route, then set and seal it
	ibcRouter := porttypes.NewRouter()
	ibcRouter.AddRoute(ibctransfertypes.ModuleName, transferStack)
	ibcRouter.AddRoute(wasmdTypes.ModuleName, wasmStack)
	ibcRouter.AddRoute(icahosttypes.SubModuleName, icaHostStack)
	app.Keepers.IBCKeeper.SetRouter(ibcRouter)

	app.Keepers.GovKeeper = *govkeeper.NewKeeper(
		appCodec,
		keys[govtypes.StoreKey],
		app.Keepers.AccountKeeper,
		app.Keepers.BankKeeper,
		app.Keepers.StakingKeeper,
		app.MsgServiceRouter(),
		govtypes.DefaultConfig(),
		govModuleAddr,
	)
	app.Keepers.GovKeeper.SetLegacyRouter(govRouter)
	/****  Module Options ****/

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	skipGenesisInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	app.mm = module.NewManager(
		genutil.NewAppModule(
			app.Keepers.AccountKeeper,
			app.Keepers.StakingKeeper,
			app.BaseApp.DeliverTx,
			encodingConfig.TxConfig,
		),
		auth.NewAppModule(appCodec, app.Keepers.AccountKeeper, nil, app.getSubspace(authtypes.ModuleName)),
		vesting.NewAppModule(app.Keepers.AccountKeeper, app.Keepers.BankKeeper),
		bank.NewAppModule(appCodec, app.Keepers.BankKeeper, app.Keepers.AccountKeeper, app.getSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.Keepers.CapabilityKeeper, false),
		gov.NewAppModule(appCodec, &app.Keepers.GovKeeper, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, app.getSubspace(govtypes.ModuleName)),
		groupmodule.NewAppModule(appCodec, app.Keepers.GroupKeeper, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, app.interfaceRegistry),
		nftmodule.NewAppModule(appCodec, app.Keepers.NFTKeeper, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, app.interfaceRegistry),
		mint.NewAppModule(appCodec, app.Keepers.MintKeeper, app.Keepers.AccountKeeper, nil, app.getSubspace(minttypes.ModuleName)),
		slashing.NewAppModule(appCodec, app.Keepers.SlashingKeeper, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, app.Keepers.StakingKeeper, app.getSubspace(slashingtypes.ModuleName)),
		distr.NewAppModule(appCodec, app.Keepers.DistrKeeper, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, app.Keepers.StakingKeeper, app.getSubspace(distrtypes.ModuleName)),
		staking.NewAppModule(appCodec, app.Keepers.StakingKeeper, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, app.getSubspace(stakingtypes.ModuleName)),
		upgrade.NewAppModule(&app.Keepers.UpgradeKeeper),
		wasm.NewAppModule(appCodec, &app.Keepers.WASMKeeper, app.Keepers.StakingKeeper, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, app.MsgServiceRouter(), app.getSubspace(wasmdTypes.ModuleName)),
		evidence.NewAppModule(app.Keepers.EvidenceKeeper),
		feegrantmodule.NewAppModule(appCodec, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, app.Keepers.FeeGrantKeeper, app.interfaceRegistry),
		authzmodule.NewAppModule(appCodec, app.Keepers.AuthzKeeper, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, app.interfaceRegistry),
		ibc.NewAppModule(app.Keepers.IBCKeeper),
		params.NewAppModule(app.Keepers.ParamsKeeper),
		transferModule,
		ibcfee.NewAppModule(app.Keepers.IBCFeeKeeper),
		ica.NewAppModule(nil, &app.Keepers.ICAHostKeeper),
		consensus.NewAppModule(appCodec, app.Keepers.ConsensusParamsKeeper),
		tracking.NewAppModule(app.appCodec, app.Keepers.TrackingKeeper),
		rewards.NewAppModule(app.appCodec, app.Keepers.RewardsKeeper),
		genmsg.NewAppModule(app.MsgServiceRouter()),
		crisis.NewAppModule(&app.Keepers.CrisisKeeper, skipGenesisInvariants, app.getSubspace(crisistypes.ModuleName)), // always be last to make sure that it checks for all invariants and not only part of them
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	app.mm.SetOrderBeginBlockers(
		upgradetypes.ModuleName,
		capabilitytypes.ModuleName,
		minttypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		govtypes.ModuleName,
		nft.ModuleName,
		crisistypes.ModuleName, // doesn't have BeginBlocker, so order is not important
		genutiltypes.ModuleName,
		genmsg.ModuleName,
		group.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		vestingtypes.ModuleName,
		consensusparamtypes.ModuleName,
		// additional non simd modules
		ibcexported.ModuleName,
		ibctransfertypes.ModuleName,
		ibcfeetypes.ModuleName,
		icatypes.ModuleName,
		// wasm
		wasmdTypes.ModuleName,
		// wasm gas tracking
		trackingTypes.ModuleName,
		rewardsTypes.ModuleName,
	)

	app.mm.SetOrderEndBlockers(
		// we have to specify all modules here (Cosmos's order is taken as a reference)
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		ibctransfertypes.ModuleName,
		ibcfeetypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		feegrant.ModuleName,
		authz.ModuleName,
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		nft.ModuleName,
		slashingtypes.ModuleName,
		minttypes.ModuleName,
		genutiltypes.ModuleName,
		genmsg.ModuleName,
		group.ModuleName,
		evidencetypes.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		consensusparamtypes.ModuleName,
		// wasm
		wasmdTypes.ModuleName,
		// wasm gas tracking
		trackingTypes.ModuleName,
		rewardsTypes.ModuleName,
		// invariants checks are always the last to run
		crisistypes.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	// NOTE: wasm module should be at the end as it can call other module functionality direct or via message dispatching during
	// genesis phase. For example bank transfer, auth account check, staking, ...
	app.mm.SetOrderInitGenesis(
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		nft.ModuleName,
		minttypes.ModuleName,
		rewardsTypes.ModuleName,
		genutiltypes.ModuleName,
		group.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		consensusparamtypes.ModuleName,
		// additional non simd modules
		ibcexported.ModuleName,
		ibctransfertypes.ModuleName,
		ibcfeetypes.ModuleName,
		icatypes.ModuleName,
		// wasm after ibc transfer
		wasmdTypes.ModuleName,
		// wasm gas tracking
		trackingTypes.ModuleName,
		genmsg.ModuleName,
		// invariants checks are always the last to run
		crisistypes.ModuleName,
	)

	// Uncomment if you want to set a custom migration order here.
	// app.mm.SetOrderMigrations(custom order)

	app.mm.RegisterInvariants(&app.Keepers.CrisisKeeper)

	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(app.configurator)
	app.setupUpgrades()

	// create the simulation manager and define the order of the modules for deterministic simulations
	//
	// NOTE: this is not required apps that don't use the simulator for fuzz testing
	// transactions
	app.sm = module.NewSimulationManager(
		auth.NewAppModule(appCodec, app.Keepers.AccountKeeper, authsims.RandomGenesisAccounts, app.getSubspace(authtypes.ModuleName)),
		bank.NewAppModule(appCodec, app.Keepers.BankKeeper, app.Keepers.AccountKeeper, app.getSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.Keepers.CapabilityKeeper, false),
		feegrantmodule.NewAppModule(appCodec, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, app.Keepers.FeeGrantKeeper, app.interfaceRegistry),
		authzmodule.NewAppModule(appCodec, app.Keepers.AuthzKeeper, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, app.interfaceRegistry),
		gov.NewAppModule(appCodec, &app.Keepers.GovKeeper, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, app.getSubspace(govtypes.ModuleName)),
		mint.NewAppModule(appCodec, app.Keepers.MintKeeper, app.Keepers.AccountKeeper, nil, app.getSubspace(minttypes.ModuleName)),
		staking.NewAppModule(appCodec, app.Keepers.StakingKeeper, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, app.getSubspace(stakingtypes.ModuleName)),
		distr.NewAppModule(appCodec, app.Keepers.DistrKeeper, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, app.Keepers.StakingKeeper, app.getSubspace(distrtypes.ModuleName)),
		slashing.NewAppModule(appCodec, app.Keepers.SlashingKeeper, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, app.Keepers.StakingKeeper, app.getSubspace(slashingtypes.ModuleName)),
		params.NewAppModule(app.Keepers.ParamsKeeper),
		evidence.NewAppModule(app.Keepers.EvidenceKeeper),
		wasm.NewAppModule(appCodec, &app.Keepers.WASMKeeper, app.Keepers.StakingKeeper, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, app.MsgServiceRouter(), app.getSubspace(wasmdTypes.ModuleName)),
		ibc.NewAppModule(app.Keepers.IBCKeeper),
		transferModule,
	)

	app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	anteHandler, err := NewAnteHandler(
		HandlerOptions{
			HandlerOptions: ante.HandlerOptions{
				AccountKeeper:   app.Keepers.AccountKeeper,
				BankKeeper:      app.Keepers.BankKeeper,
				FeegrantKeeper:  app.Keepers.FeeGrantKeeper,
				SignModeHandler: encodingConfig.TxConfig.SignModeHandler(),
				SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
			},
			IBCKeeper:             app.Keepers.IBCKeeper,
			WasmConfig:            &wasmConfig,
			RewardsAnteBankKeeper: app.Keepers.BankKeeper,
			TXCounterStoreKey:     keys[wasmdTypes.StoreKey],
			TrackingKeeper:        app.Keepers.TrackingKeeper,
			RewardsKeeper:         app.Keepers.RewardsKeeper,
		},
	)
	if err != nil {
		panic(fmt.Errorf("failed to create AnteHandler: %s", err))
	}
	postHandler, err := posthandler.NewPostHandler(
		posthandler.HandlerOptions{},
	)
	if err != nil {
		panic(fmt.Errorf("failed to create PostHandler: %s", err))
	}

	app.SetAnteHandler(anteHandler)
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)
	app.SetPostHandler(postHandler)

	autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(app.mm.Modules))

	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		panic(err)
	}
	reflectionv1.RegisterReflectionServiceServer(app.GRPCQueryRouter(), reflectionSvc)

	// Register snapshot extensions to enable state-sync for wasm - must be before Loading version
	if manager := app.SnapshotManager(); manager != nil {
		err := manager.RegisterExtensions(
			wasmdKeeper.NewWasmSnapshotter(app.CommitMultiStore(), &app.Keepers.WASMKeeper),
		)
		if err != nil {
			panic(fmt.Errorf("failed to register snapshot extension: %s", err))
		}
	}

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(fmt.Sprintf("failed to load latest version: %s", err))
		}
		ctx := app.BaseApp.NewUncachedContext(true, tmproto.Header{})

		// Initialize pinned codes in wasmvm as they are not persisted there
		if err := app.Keepers.WASMKeeper.InitializePinnedCodes(ctx); err != nil {
			tmos.Exit(fmt.Sprintf("failed initialize pinned codes %s", err))
		}
	}

	app.ScopedIBCKeeper = scopedIBCKeeper
	app.ScopedTransferKeeper = scopedTransferKeeper
	app.ScopedWASMKeeper = scopedWasmKeeper
	app.ScopedICAHostKeeper = scopedICAHostKeeper
	return app
}

// Name returns the name of the App
func (app *ArchwayApp) Name() string { return app.BaseApp.Name() }

// BeginBlocker processes application updates every begin block
func (app *ArchwayApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// EndBlocker application updates every end block
func (app *ArchwayApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// InitChainer application update at chain initialization
func (app *ArchwayApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}

	app.Keepers.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())

	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height
func (app *ArchwayApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *ArchwayApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// LegacyAmino returns legacy amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *ArchwayApp) LegacyAmino() *codec.LegacyAmino { //nolint:staticcheck
	return app.legacyAmino
}

// getSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *ArchwayApp) getSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.Keepers.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// SimulationManager implements the SimulationApp interface
func (app *ArchwayApp) SimulationManager() *module.SimulationManager {
	return app.sm
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *ArchwayApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register node gRPC service for grpc-gateway.
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	if err := server.RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *ArchwayApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *ArchwayApp) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(clientCtx, app.BaseApp.GRPCQueryRouter(), app.interfaceRegistry, app.Query)
}

// RegisterNodeService implements the Application.RegisterNodeService method.
func (app *ArchwayApp) RegisterNodeService(clientCtx client.Context) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter())
}

func (app *ArchwayApp) AppCodec() codec.Codec {
	return app.appCodec
}

// GetMaccPerms returns a copy of the module account permissions
func GetMaccPerms() map[string][]string {
	dupMaccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		dupMaccPerms[k] = v
	}
	return dupMaccPerms
}

// BlockedAddresses returns all the app's blocked account addresses.
func BlockedAddresses() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range GetMaccPerms() {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	// allow the following addresses to receive funds
	delete(modAccAddrs, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	return modAccAddrs
}

func GetGovProposalHandlers() []govclient.ProposalHandler {
	var govProposalHandlers []govclient.ProposalHandler

	govProposalHandlers = append(govProposalHandlers,
		paramsclient.ProposalHandler,
		upgradeclient.LegacyProposalHandler,
		upgradeclient.LegacyCancelProposalHandler,
		ibcclientclient.UpdateClientProposalHandler,
		ibcclientclient.UpgradeProposalHandler,
	)

	return govProposalHandlers
}

// initParamsKeeper init params keeper and its subspaces
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(minttypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName)
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(ibcexported.ModuleName)
	paramsKeeper.Subspace(icahosttypes.SubModuleName)
	paramsKeeper.Subspace(wasmdTypes.ModuleName)
	paramsKeeper.Subspace(rewardsTypes.ModuleName)

	return paramsKeeper
}
