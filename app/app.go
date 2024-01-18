package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/archway-network/archway/app/keepers"
	"github.com/archway-network/archway/x/cwfees"
	"github.com/archway-network/archway/x/genmsg"

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

	"github.com/archway-network/archway/x/callback"
	callbackkeeper "github.com/archway-network/archway/x/callback/keeper"
	callbackTypes "github.com/archway-network/archway/x/callback/types"

	"github.com/archway-network/archway/x/rewards"
	rewardskeeper "github.com/archway-network/archway/x/rewards/keeper"
	"github.com/archway-network/archway/x/rewards/mintbankkeeper"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"

	"github.com/archway-network/archway/x/tracking"
	trackingKeeper "github.com/archway-network/archway/x/tracking/keeper"
	trackingTypes "github.com/archway-network/archway/x/tracking/types"

	"github.com/CosmWasm/wasmd/x/wasm"

	archwayappparams "github.com/archway-network/archway/app/params"
	archway "github.com/archway-network/archway/types"
)

const AppName = "Archway"

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
		callback.AppModuleBasic{},
		cwfees.AppModule{},
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
		callbackTypes.ModuleName:             nil,
	}
)

var (
	_ runtime.AppI            = (*ArchwayApp)(nil)
	_ servertypes.Application = (*ArchwayApp)(nil)
)

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, "."+NodeDir)

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

	// keepers
	Keepers       keepers.ArchwayKeepers
	ScopedKeepers keepers.ArchwayScopedKeepers

	// the module manager
	mm *module.Manager

	// simulation manager
	sm *module.SimulationManager

	// module configurator
	configurator module.Configurator

	// authority address
	authority string
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
	appCodec, legacyAmino, interfaceRegistry := encodingConfig.Marshaler, encodingConfig.Amino, encodingConfig.InterfaceRegistry

	bApp := baseapp.NewBaseApp(AppName, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	app := &ArchwayApp{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
		keys:              make(map[string]*storetypes.KVStoreKey),
		tkeys:             make(map[string]*storetypes.TransientStoreKey),
		memKeys:           make(map[string]*storetypes.MemoryStoreKey),
		Keepers:           keepers.ArchwayKeepers{},
		ScopedKeepers:     keepers.ArchwayScopedKeepers{},
		authority:         authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	}
	modules := make([]module.AppModule, 0)
	simModules := make([]module.AppModuleSimulation, 0)
	beginBlockerModules := make([]string, 0)
	endBlockerModules := make([]string, 0)
	initGenesisModules := make([]string, 0)

	//
	// BEGIN: Register Cosmos SDK modules
	//

	// NOTE: capability module must occur first in genesis init so that so that other modules
	// that want to create or claim capabilities afterwards in InitChain can do so safely.
	initGenesisModules = append(initGenesisModules, capabilitytypes.ModuleName)

	// 'auth' module
	app.keys[authtypes.StoreKey] = storetypes.NewKVStoreKey(authtypes.StoreKey)
	app.Keepers.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		app.keys[authtypes.StoreKey],
		authtypes.ProtoBaseAccount,
		maccPerms,
		Bech32Prefix,
		app.authority,
	)
	modules = append(modules, auth.NewAppModule(appCodec, app.Keepers.AccountKeeper, nil, app.getSubspace(authtypes.ModuleName)))
	simModules = append(simModules, auth.NewAppModule(appCodec, app.Keepers.AccountKeeper, authsims.RandomGenesisAccounts, app.getSubspace(authtypes.ModuleName)))
	beginBlockerModules = append(beginBlockerModules, authtypes.ModuleName)
	endBlockerModules = append(endBlockerModules, authtypes.ModuleName)
	initGenesisModules = append(initGenesisModules, authtypes.ModuleName)

	// 'bank' module
	app.keys[banktypes.StoreKey] = storetypes.NewKVStoreKey(banktypes.StoreKey)
	app.Keepers.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		app.keys[banktypes.StoreKey],
		app.Keepers.AccountKeeper,
		BlockedAddresses(),
		app.authority,
	)
	modules = append(modules, bank.NewAppModule(appCodec, app.Keepers.BankKeeper, app.Keepers.AccountKeeper, app.getSubspace(banktypes.ModuleName)))
	simModules = append(simModules, bank.NewAppModule(appCodec, app.Keepers.BankKeeper, app.Keepers.AccountKeeper, app.getSubspace(banktypes.ModuleName)))
	beginBlockerModules = append(beginBlockerModules, banktypes.ModuleName)
	endBlockerModules = append(endBlockerModules, banktypes.ModuleName)
	initGenesisModules = append(initGenesisModules, banktypes.ModuleName)

	// 'authz' module
	app.keys[authzkeeper.StoreKey] = storetypes.NewKVStoreKey(authzkeeper.StoreKey)
	app.Keepers.AuthzKeeper = authzkeeper.NewKeeper(
		app.keys[authzkeeper.StoreKey],
		appCodec,
		app.MsgServiceRouter(),
		app.Keepers.AccountKeeper,
	)
	modules = append(modules, authzmodule.NewAppModule(appCodec, app.Keepers.AuthzKeeper, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, interfaceRegistry))
	simModules = append(simModules, authzmodule.NewAppModule(appCodec, app.Keepers.AuthzKeeper, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, app.interfaceRegistry))
	beginBlockerModules = append(beginBlockerModules, authz.ModuleName)
	endBlockerModules = append(endBlockerModules, authz.ModuleName)
	initGenesisModules = append(initGenesisModules, authz.ModuleName)

	// 'capability' module
	app.keys[capabilitytypes.StoreKey] = storetypes.NewKVStoreKey(capabilitytypes.StoreKey)
	app.memKeys[capabilitytypes.MemStoreKey] = storetypes.NewMemoryStoreKey(capabilitytypes.MemStoreKey)
	app.Keepers.CapabilityKeeper = capabilitykeeper.NewKeeper(
		appCodec,
		app.keys[capabilitytypes.StoreKey],
		app.memKeys[capabilitytypes.MemStoreKey],
	)
	modules = append(modules, capability.NewAppModule(appCodec, *app.Keepers.CapabilityKeeper, false))
	simModules = append(simModules, capability.NewAppModule(appCodec, *app.Keepers.CapabilityKeeper, false))
	beginBlockerModules = append(beginBlockerModules, capabilitytypes.ModuleName)
	endBlockerModules = append(endBlockerModules, capabilitytypes.ModuleName)

	// 'consensus' module
	app.keys[consensusparamtypes.StoreKey] = storetypes.NewKVStoreKey(consensusparamtypes.StoreKey)
	app.Keepers.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(
		appCodec,
		app.keys[consensusparamtypes.StoreKey],
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.SetParamStore(&app.Keepers.ConsensusParamsKeeper)
	modules = append(modules, consensus.NewAppModule(appCodec, app.Keepers.ConsensusParamsKeeper))
	beginBlockerModules = append(beginBlockerModules, consensusparamtypes.ModuleName)
	endBlockerModules = append(endBlockerModules, consensusparamtypes.ModuleName)
	initGenesisModules = append(initGenesisModules, consensusparamtypes.ModuleName)

	// 'crisis' module
	app.keys[crisistypes.StoreKey] = storetypes.NewKVStoreKey(crisistypes.StoreKey)
	app.Keepers.CrisisKeeper = *crisiskeeper.NewKeeper(
		appCodec,
		app.keys[crisistypes.StoreKey],
		invCheckPeriod,
		app.Keepers.BankKeeper,
		authtypes.FeeCollectorName,
		app.authority,
	)
	beginBlockerModules = append(beginBlockerModules, crisistypes.ModuleName)
	// NOTE: crisis endblocker is added at the end as invariant checks are always last to run
	// NOTE: crisis init genesis is added at the end as invariant checks are always last to run

	// 'feegrant' module
	app.keys[feegrant.StoreKey] = storetypes.NewKVStoreKey(feegrant.StoreKey)
	app.Keepers.FeeGrantKeeper = feegrantkeeper.NewKeeper(
		appCodec,
		app.keys[feegrant.StoreKey],
		app.Keepers.AccountKeeper,
	)
	modules = append(modules, feegrantmodule.NewAppModule(appCodec, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, app.Keepers.FeeGrantKeeper, interfaceRegistry))
	simModules = append(simModules, feegrantmodule.NewAppModule(appCodec, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, app.Keepers.FeeGrantKeeper, app.interfaceRegistry))
	beginBlockerModules = append(beginBlockerModules, feegrant.ModuleName)
	endBlockerModules = append(endBlockerModules, feegrant.ModuleName)
	initGenesisModules = append(initGenesisModules, feegrant.ModuleName)

	// 'group' module
	app.keys[group.StoreKey] = storetypes.NewKVStoreKey(group.StoreKey)
	app.Keepers.GroupKeeper = groupkeeper.NewKeeper(
		app.keys[group.StoreKey],
		appCodec,
		app.MsgServiceRouter(),
		app.Keepers.AccountKeeper,
		group.DefaultConfig(),
	)
	modules = append(modules, groupmodule.NewAppModule(appCodec, app.Keepers.GroupKeeper, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, interfaceRegistry))
	beginBlockerModules = append(beginBlockerModules, group.ModuleName)
	endBlockerModules = append(endBlockerModules, group.ModuleName)
	initGenesisModules = append(initGenesisModules, group.ModuleName)

	// 'staking' module
	app.keys[stakingtypes.StoreKey] = storetypes.NewKVStoreKey(stakingtypes.StoreKey)
	app.Keepers.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec,
		app.keys[stakingtypes.StoreKey],
		app.Keepers.AccountKeeper,
		app.Keepers.BankKeeper,
		app.authority,
	)
	stakingHooks := make([]stakingtypes.StakingHooks, 0)
	modules = append(modules, staking.NewAppModule(appCodec, app.Keepers.StakingKeeper, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, app.getSubspace(stakingtypes.ModuleName)))
	simModules = append(simModules, staking.NewAppModule(appCodec, app.Keepers.StakingKeeper, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, app.getSubspace(stakingtypes.ModuleName)))
	beginBlockerModules = append(beginBlockerModules, stakingtypes.ModuleName)
	endBlockerModules = append(endBlockerModules, stakingtypes.ModuleName)
	initGenesisModules = append(initGenesisModules, stakingtypes.ModuleName)

	// 'nft' module
	app.keys[nftkeeper.StoreKey] = storetypes.NewKVStoreKey(nftkeeper.StoreKey)
	app.Keepers.NFTKeeper = nftkeeper.NewKeeper(
		app.keys[nftkeeper.StoreKey],
		appCodec,
		app.Keepers.AccountKeeper,
		app.Keepers.BankKeeper,
	)
	modules = append(modules, nftmodule.NewAppModule(appCodec, app.Keepers.NFTKeeper, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, interfaceRegistry))
	beginBlockerModules = append(beginBlockerModules, nft.ModuleName)
	endBlockerModules = append(endBlockerModules, nft.ModuleName)
	initGenesisModules = append(initGenesisModules, nft.ModuleName)

	// 'gov' module - depends on
	app.keys[govtypes.StoreKey] = storetypes.NewKVStoreKey(govtypes.StoreKey)
	app.Keepers.GovKeeper = *govkeeper.NewKeeper(
		appCodec,
		app.keys[govtypes.StoreKey],
		app.Keepers.AccountKeeper,
		app.Keepers.BankKeeper,
		app.Keepers.StakingKeeper,
		app.MsgServiceRouter(),
		govtypes.DefaultConfig(),
		app.authority,
	)
	// Set legacy router for backwards compatibility with gov v1beta1
	govLegacyRouter := govV1Beta1types.NewRouter()
	govLegacyRouter.AddRoute(govtypes.RouterKey, govV1Beta1types.ProposalHandler)
	modules = append(modules, gov.NewAppModule(appCodec, &app.Keepers.GovKeeper, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, app.getSubspace(govtypes.ModuleName)))
	simModules = append(simModules, gov.NewAppModule(appCodec, &app.Keepers.GovKeeper, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, app.getSubspace(govtypes.ModuleName)))
	beginBlockerModules = append(beginBlockerModules, govtypes.ModuleName)
	endBlockerModules = append(endBlockerModules, govtypes.ModuleName)
	initGenesisModules = append(initGenesisModules, govtypes.ModuleName)

	// 'distribution' module
	app.keys[distrtypes.StoreKey] = storetypes.NewKVStoreKey(distrtypes.StoreKey)
	app.Keepers.DistrKeeper = distrkeeper.NewKeeper(
		appCodec,
		app.keys[distrtypes.StoreKey],
		app.Keepers.AccountKeeper,
		app.Keepers.BankKeeper,
		app.Keepers.StakingKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	stakingHooks = append(stakingHooks, app.Keepers.DistrKeeper.Hooks())
	modules = append(modules, distr.NewAppModule(appCodec, app.Keepers.DistrKeeper, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, app.Keepers.StakingKeeper, app.getSubspace(distrtypes.ModuleName)))
	simModules = append(simModules, distr.NewAppModule(appCodec, app.Keepers.DistrKeeper, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, app.Keepers.StakingKeeper, app.getSubspace(distrtypes.ModuleName)))
	beginBlockerModules = append(beginBlockerModules, distrtypes.ModuleName)
	endBlockerModules = append(endBlockerModules, distrtypes.ModuleName)
	initGenesisModules = append(initGenesisModules, distrtypes.ModuleName)

	// 'slashing' module
	app.keys[slashingtypes.StoreKey] = storetypes.NewKVStoreKey(slashingtypes.StoreKey)
	app.Keepers.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec,
		encodingConfig.Amino,
		app.keys[slashingtypes.StoreKey],
		app.Keepers.StakingKeeper,
		app.authority,
	)
	stakingHooks = append(stakingHooks, app.Keepers.SlashingKeeper.Hooks())
	modules = append(modules, slashing.NewAppModule(appCodec, app.Keepers.SlashingKeeper, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, app.Keepers.StakingKeeper, app.getSubspace(slashingtypes.ModuleName)))
	simModules = append(simModules, slashing.NewAppModule(appCodec, app.Keepers.SlashingKeeper, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, app.Keepers.StakingKeeper, app.getSubspace(slashingtypes.ModuleName)))
	beginBlockerModules = append(beginBlockerModules, slashingtypes.ModuleName)
	endBlockerModules = append(endBlockerModules, slashingtypes.ModuleName)
	initGenesisModules = append(initGenesisModules, slashingtypes.ModuleName)

	// 'params' module
	app.keys[paramstypes.StoreKey] = storetypes.NewKVStoreKey(paramstypes.StoreKey)
	app.tkeys[paramstypes.TStoreKey] = storetypes.NewTransientStoreKey(paramstypes.TStoreKey)
	app.Keepers.ParamsKeeper = initParamsKeeper(
		appCodec,
		legacyAmino,
		app.keys[paramstypes.StoreKey],
		app.tkeys[paramstypes.TStoreKey],
	)
	govLegacyRouter.AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.Keepers.ParamsKeeper))
	modules = append(modules, params.NewAppModule(app.Keepers.ParamsKeeper))
	simModules = append(simModules, params.NewAppModule(app.Keepers.ParamsKeeper))
	beginBlockerModules = append(beginBlockerModules, paramstypes.ModuleName)
	endBlockerModules = append(endBlockerModules, paramstypes.ModuleName)
	initGenesisModules = append(initGenesisModules, paramstypes.ModuleName)

	// 'evidence' module
	app.keys[evidencetypes.StoreKey] = storetypes.NewKVStoreKey(evidencetypes.StoreKey)
	app.Keepers.EvidenceKeeper = *evidencekeeper.NewKeeper(
		appCodec,
		app.keys[evidencetypes.StoreKey],
		app.Keepers.StakingKeeper,
		app.Keepers.SlashingKeeper,
	)
	modules = append(modules, evidence.NewAppModule(app.Keepers.EvidenceKeeper))
	simModules = append(simModules, evidence.NewAppModule(app.Keepers.EvidenceKeeper))
	beginBlockerModules = append(beginBlockerModules, evidencetypes.ModuleName)
	endBlockerModules = append(endBlockerModules, evidencetypes.ModuleName)
	initGenesisModules = append(initGenesisModules, evidencetypes.ModuleName)

	// 'upgrade' module
	app.keys[upgradetypes.StoreKey] = storetypes.NewKVStoreKey(upgradetypes.StoreKey)
	app.Keepers.UpgradeKeeper = *upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		app.keys[upgradetypes.StoreKey],
		appCodec,
		homePath,
		app.BaseApp,
		app.authority,
	)
	govLegacyRouter.AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(&app.Keepers.UpgradeKeeper))
	modules = append(modules, upgrade.NewAppModule(&app.Keepers.UpgradeKeeper))
	beginBlockerModules = append(beginBlockerModules, upgradetypes.ModuleName)
	endBlockerModules = append(endBlockerModules, upgradetypes.ModuleName)
	initGenesisModules = append(initGenesisModules, upgradetypes.ModuleName)

	// 'genutil' module
	modules = append(modules, genutil.NewAppModule(app.Keepers.AccountKeeper, app.Keepers.StakingKeeper, app.BaseApp.DeliverTx, encodingConfig.TxConfig))
	beginBlockerModules = append(beginBlockerModules, genutiltypes.ModuleName)
	endBlockerModules = append(endBlockerModules, genutiltypes.ModuleName)

	// 'vesting' module
	modules = append(modules, vesting.NewAppModule(app.Keepers.AccountKeeper, app.Keepers.BankKeeper))
	beginBlockerModules = append(beginBlockerModules, vestingtypes.ModuleName)
	endBlockerModules = append(endBlockerModules, vestingtypes.ModuleName)
	initGenesisModules = append(initGenesisModules, vestingtypes.ModuleName)

	//
	// END: Register Cosmos SDK modules
	//

	//
	// BEGIN: Register IBC modules
	//

	// 'ibc' module
	app.keys[ibcexported.StoreKey] = storetypes.NewKVStoreKey(ibcexported.StoreKey)
	app.ScopedKeepers.ScopedIBCKeeper = app.Keepers.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	app.Keepers.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		app.keys[ibcexported.StoreKey],
		app.getSubspace(ibcexported.ModuleName),
		app.Keepers.StakingKeeper,
		app.Keepers.UpgradeKeeper,
		app.ScopedKeepers.ScopedIBCKeeper,
	)
	govLegacyRouter.AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(app.Keepers.IBCKeeper.ClientKeeper))
	modules = append(modules, ibc.NewAppModule(app.Keepers.IBCKeeper))
	simModules = append(simModules, ibc.NewAppModule(app.Keepers.IBCKeeper))
	beginBlockerModules = append(beginBlockerModules, ibcexported.ModuleName)
	endBlockerModules = append(endBlockerModules, ibcexported.ModuleName)
	initGenesisModules = append(initGenesisModules, ibcexported.ModuleName)

	// 'ibcfeekeeper' module
	app.keys[ibcfeetypes.StoreKey] = storetypes.NewKVStoreKey(ibcfeetypes.StoreKey)
	app.Keepers.IBCFeeKeeper = ibcfeekeeper.NewKeeper(
		appCodec,
		app.keys[ibcfeetypes.StoreKey],
		app.Keepers.IBCKeeper.ChannelKeeper, // may be replaced with IBC middleware
		app.Keepers.IBCKeeper.ChannelKeeper,
		&app.Keepers.IBCKeeper.PortKeeper,
		app.Keepers.AccountKeeper,
		app.Keepers.BankKeeper,
	)
	modules = append(modules, ibcfee.NewAppModule(app.Keepers.IBCFeeKeeper))
	beginBlockerModules = append(beginBlockerModules, ibcfeetypes.ModuleName)
	endBlockerModules = append(endBlockerModules, ibcfeetypes.ModuleName)
	initGenesisModules = append(initGenesisModules, ibcfeetypes.ModuleName)

	// 'ibctransfer' module
	app.keys[ibctransfertypes.StoreKey] = storetypes.NewKVStoreKey(ibctransfertypes.StoreKey)
	app.ScopedKeepers.ScopedTransferKeeper = app.Keepers.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	app.Keepers.TransferKeeper = ibctransferkeeper.NewKeeper(
		appCodec,
		app.keys[ibctransfertypes.StoreKey],
		app.getSubspace(ibctransfertypes.ModuleName),
		app.Keepers.IBCFeeKeeper,
		app.Keepers.IBCKeeper.ChannelKeeper,
		&app.Keepers.IBCKeeper.PortKeeper,
		app.Keepers.AccountKeeper,
		app.Keepers.BankKeeper,
		app.ScopedKeepers.ScopedTransferKeeper,
	)
	modules = append(modules, transfer.NewAppModule(app.Keepers.TransferKeeper))
	simModules = append(simModules, transfer.NewAppModule(app.Keepers.TransferKeeper))
	beginBlockerModules = append(beginBlockerModules, ibctransfertypes.ModuleName)
	endBlockerModules = append(endBlockerModules, ibctransfertypes.ModuleName)
	initGenesisModules = append(initGenesisModules, ibctransfertypes.ModuleName)

	// 'icahost' module
	app.keys[icahosttypes.StoreKey] = storetypes.NewKVStoreKey(icahosttypes.StoreKey)
	app.ScopedKeepers.ScopedICAHostKeeper = app.Keepers.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	app.Keepers.ICAHostKeeper = icahostkeeper.NewKeeper(
		appCodec,
		app.keys[icahosttypes.StoreKey],
		app.getSubspace(icahosttypes.SubModuleName),
		app.Keepers.IBCFeeKeeper,
		app.Keepers.IBCKeeper.ChannelKeeper,
		&app.Keepers.IBCKeeper.PortKeeper,
		app.Keepers.AccountKeeper,
		app.ScopedKeepers.ScopedICAHostKeeper,
		app.MsgServiceRouter(),
	)
	modules = append(modules, ica.NewAppModule(nil, &app.Keepers.ICAHostKeeper))
	beginBlockerModules = append(beginBlockerModules, icatypes.ModuleName)
	endBlockerModules = append(endBlockerModules, icatypes.ModuleName)
	initGenesisModules = append(initGenesisModules, icatypes.ModuleName)

	var transferStack porttypes.IBCModule
	transferStack = transfer.NewIBCModule(app.Keepers.TransferKeeper)
	transferStack = ibcfee.NewIBCMiddleware(transferStack, app.Keepers.IBCFeeKeeper)

	// Create Interchain Accounts Stack
	// RecvPacket, message that originates from core IBC and goes down to app, the flow is:
	// channel.RecvPacket -> fee.OnRecvPacket -> icaHost.OnRecvPacket
	var icaHostStack porttypes.IBCModule
	icaHostStack = icahost.NewIBCModule(app.Keepers.ICAHostKeeper)
	icaHostStack = ibcfee.NewIBCMiddleware(icaHostStack, app.Keepers.IBCFeeKeeper)
	// Create fee enabled wasm ibc Stack
	var wasmStack porttypes.IBCModule
	wasmStack = wasm.NewIBCHandler(app.Keepers.WASMKeeper, app.Keepers.IBCKeeper.ChannelKeeper, app.Keepers.IBCFeeKeeper)
	wasmStack = ibcfee.NewIBCMiddleware(wasmStack, app.Keepers.IBCFeeKeeper)

	// create static IBC router, add transfer route, add wasm route, then set and seal it
	ibcRouter := porttypes.NewRouter()
	ibcRouter.AddRoute(ibctransfertypes.ModuleName, transferStack)
	ibcRouter.AddRoute(wasmdTypes.ModuleName, wasmStack)
	ibcRouter.AddRoute(icahosttypes.SubModuleName, icaHostStack)
	app.Keepers.IBCKeeper.SetRouter(ibcRouter)

	//
	// END: Register IBC modules
	//

	//
	// BEGIN: Register Wasmd module
	//

	// 'wasm' module
	wasmDir := filepath.Join(homePath, "wasm")
	wasmConfig, err := wasm.ReadWasmConfig(appOpts)
	if err != nil {
		panic(fmt.Sprintf("error while reading wasm config: %s", err))
	}

	supportedFeatures := "iterator,staking,stargate,cosmwasm_1_1,cosmwasm_1_2,cosmwasm_1_3,cosmwasm_1_4"
	defaultGasRegister := wasmdTypes.NewDefaultWasmGasRegister()
	wasmer, err := cosmwasm.NewVM(filepath.Join(wasmDir, "wasm"), supportedFeatures, 32, wasmConfig.ContractDebugMode, wasmConfig.MemoryCacheSize)
	if err != nil {
		panic(err)
	}
	trackingWasmVm := wasmdTypes.NewTrackingWasmerEngine(wasmer, &wasmdTypes.NoOpContractGasProcessor{})
	wasmOpts = append(wasmOpts, wasmdKeeper.WithWasmEngine(trackingWasmVm), wasmdKeeper.WithGasRegister(defaultGasRegister))
	// Archway specific options (using a pointer as the keeper is post-initialized below)
	wasmOpts = append(wasmOpts, wasmbinding.BuildWasmOptions(&app.Keepers.RewardsKeeper, &app.Keepers.GovKeeper)...)

	app.keys[wasmdTypes.StoreKey] = storetypes.NewKVStoreKey(wasmdTypes.StoreKey)
	app.ScopedKeepers.ScopedWASMKeeper = app.Keepers.CapabilityKeeper.ScopeToModule(wasmdTypes.ModuleName)
	app.Keepers.WASMKeeper = wasmdKeeper.NewKeeper(
		appCodec,
		app.keys[wasmdTypes.StoreKey],
		app.Keepers.AccountKeeper,
		app.Keepers.BankKeeper,
		app.Keepers.StakingKeeper,
		distrkeeper.NewQuerier(app.Keepers.DistrKeeper),
		app.Keepers.IBCFeeKeeper, // ISC4 Wrapper: fee IBC middleware
		app.Keepers.IBCKeeper.ChannelKeeper,
		&app.Keepers.IBCKeeper.PortKeeper,
		app.ScopedKeepers.ScopedWASMKeeper,
		app.Keepers.TransferKeeper,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		wasmDir,
		wasmConfig,
		supportedFeatures,
		app.authority,
		wasmOpts...,
	)
	modules = append(modules, wasm.NewAppModule(appCodec, &app.Keepers.WASMKeeper, app.Keepers.StakingKeeper, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, app.MsgServiceRouter(), app.getSubspace(wasmdTypes.ModuleName)))
	simModules = append(simModules, wasm.NewAppModule(appCodec, &app.Keepers.WASMKeeper, app.Keepers.StakingKeeper, app.Keepers.AccountKeeper, app.Keepers.BankKeeper, app.MsgServiceRouter(), app.getSubspace(wasmdTypes.ModuleName)))
	beginBlockerModules = append(beginBlockerModules, wasmdTypes.ModuleName)
	endBlockerModules = append(endBlockerModules, wasmdTypes.ModuleName)
	initGenesisModules = append(initGenesisModules, wasmdTypes.ModuleName)
	//
	// END: Register Wasmd module
	//

	//
	// BEGIN: Register Archway modules
	//

	// 'tracking' module - tracks wasm gas usage
	app.keys[trackingTypes.StoreKey] = storetypes.NewKVStoreKey(trackingTypes.StoreKey)
	app.Keepers.TrackingKeeper = trackingKeeper.NewKeeper(
		appCodec,
		app.keys[trackingTypes.StoreKey],
		defaultGasRegister,
	)
	modules = append(modules, tracking.NewAppModule(appCodec, app.Keepers.TrackingKeeper))
	beginBlockerModules = append(beginBlockerModules, trackingTypes.ModuleName)
	endBlockerModules = append(endBlockerModules, trackingTypes.ModuleName)
	initGenesisModules = append(initGenesisModules, trackingTypes.ModuleName)
	// Setting gas recorder here to avoid cyclic loop
	trackingWasmVm.SetGasRecorder(app.Keepers.TrackingKeeper)

	// 'rewards' module
	app.keys[rewardsTypes.StoreKey] = storetypes.NewKVStoreKey(rewardsTypes.StoreKey)
	app.Keepers.RewardsKeeper = rewardskeeper.NewKeeper(
		appCodec,
		app.keys[rewardsTypes.StoreKey],
		app.Keepers.WASMKeeper,
		app.Keepers.TrackingKeeper,
		app.Keepers.AccountKeeper,
		app.Keepers.BankKeeper,
		app.getSubspace(rewardsTypes.ModuleName),
		app.authority,
	)
	modules = append(modules, rewards.NewAppModule(app.appCodec, app.Keepers.RewardsKeeper))
	beginBlockerModules = append(beginBlockerModules, rewardsTypes.ModuleName)
	endBlockerModules = append(endBlockerModules, rewardsTypes.ModuleName)
	initGenesisModules = append(initGenesisModules, rewardsTypes.ModuleName)

	// 'mint' module
	app.keys[minttypes.StoreKey] = storetypes.NewKVStoreKey(minttypes.StoreKey)
	app.Keepers.MintKeeper = mintkeeper.NewKeeper(
		appCodec,
		app.keys[minttypes.StoreKey],
		app.Keepers.StakingKeeper,
		app.Keepers.AccountKeeper,
		mintbankkeeper.NewKeeper(app.Keepers.BankKeeper, app.Keepers.RewardsKeeper),
		authtypes.FeeCollectorName,
		app.authority,
	)
	modules = append(modules, mint.NewAppModule(appCodec, app.Keepers.MintKeeper, app.Keepers.AccountKeeper, nil, app.getSubspace(minttypes.ModuleName)))
	simModules = append(simModules, mint.NewAppModule(appCodec, app.Keepers.MintKeeper, app.Keepers.AccountKeeper, nil, app.getSubspace(minttypes.ModuleName)))
	beginBlockerModules = append(beginBlockerModules, minttypes.ModuleName)
	endBlockerModules = append(endBlockerModules, minttypes.ModuleName)
	initGenesisModules = append(initGenesisModules, minttypes.ModuleName)

	// 'callback' module
	app.keys[callbackTypes.StoreKey] = storetypes.NewKVStoreKey(callbackTypes.StoreKey)
	app.Keepers.CallbackKeeper = callbackkeeper.NewKeeper(
		appCodec,
		app.keys[callbackTypes.StoreKey],
		app.Keepers.WASMKeeper,
		app.Keepers.RewardsKeeper,
		app.Keepers.BankKeeper,
		app.authority,
	)
	modules = append(modules, callback.NewAppModule(app.appCodec, app.Keepers.CallbackKeeper, app.Keepers.WASMKeeper))
	beginBlockerModules = append(beginBlockerModules, callbackTypes.ModuleName)
	endBlockerModules = append(endBlockerModules, callbackTypes.ModuleName)
	initGenesisModules = append(initGenesisModules, callbackTypes.ModuleName)

	// 'cwfees' module
	app.keys[cwfees.ModuleName] = storetypes.NewKVStoreKey(cwfees.ModuleName)
	app.Keepers.CWFeesKeeper = cwfees.NewKeeper(appCodec, app.keys[cwfees.ModuleName], app.Keepers.WASMKeeper)
	modules = append(modules, cwfees.NewAppModule(app.Keepers.CWFeesKeeper))
	beginBlockerModules = append(beginBlockerModules, cwfees.ModuleName)
	endBlockerModules = append(endBlockerModules, cwfees.ModuleName)
	initGenesisModules = append(initGenesisModules, cwfees.ModuleName)

	// 'genmsg' module
	modules = append(modules, genmsg.NewAppModule(app.MsgServiceRouter()))
	beginBlockerModules = append(beginBlockerModules, genmsg.ModuleName)
	endBlockerModules = append(endBlockerModules, genmsg.ModuleName)
	initGenesisModules = append(initGenesisModules, genmsg.ModuleName)

	//
	// END: Register Archway modules
	//

	// Wrapping up the keepers
	app.Keepers.StakingKeeper.SetHooks(stakingtypes.NewMultiStakingHooks(stakingHooks...))
	app.Keepers.GovKeeper.SetLegacyRouter(govLegacyRouter)
	app.Keepers.CapabilityKeeper.Seal()

	modules = append(modules, crisis.NewAppModule(&app.Keepers.CrisisKeeper, cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants)), app.getSubspace(crisistypes.ModuleName))) // always be last to make sure that it checks for all invariants and not only part of them
	endBlockerModules = append(endBlockerModules, crisistypes.ModuleName)                                                                                                          // NOTE: crisis endblocker is added at the end as invariant checks are always last to run
	initGenesisModules = append(initGenesisModules, genutiltypes.ModuleName)                                                                                                       //
	initGenesisModules = append(initGenesisModules, crisistypes.ModuleName)

	//
	// BEGIN: Module options
	//

	app.mm = module.NewManager(modules...)
	app.mm.RegisterInvariants(&app.Keepers.CrisisKeeper)

	app.mm.SetOrderBeginBlockers(beginBlockerModules...)
	app.mm.SetOrderEndBlockers(endBlockerModules...)
	app.mm.SetOrderInitGenesis(initGenesisModules...)
	// app.mm.SetOrderMigrations(custom order)

	app.configurator = module.NewConfigurator(appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(app.configurator)
	app.SetupUpgrades()

	// Simulation module manager
	app.sm = module.NewSimulationManager(simModules...)
	app.sm.RegisterStoreDecoders()

	//
	// END: Module options
	//

	app.MountKVStores(app.keys)
	app.MountTransientStores(app.tkeys)
	app.MountMemoryStores(app.memKeys)

	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetAnteHandler(app.AnteHandler(wasmConfig, encodingConfig))
	app.SetEndBlocker(app.EndBlocker)
	app.SetPostHandler(app.PostHandler())

	// GRPC query service
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

// AnteHandler application updates every end block
func (app *ArchwayApp) AnteHandler(wasmConfig wasmdTypes.WasmConfig, encodingConfig archwayappparams.EncodingConfig) sdk.AnteHandler {
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
			TXCounterStoreKey:     app.keys[wasmdTypes.StoreKey],
			TrackingKeeper:        app.Keepers.TrackingKeeper,
			RewardsKeeper:         app.Keepers.RewardsKeeper,
			CWFeesKeeper:          app.Keepers.CWFeesKeeper,
		},
	)
	if err != nil {
		panic(fmt.Errorf("failed to create AnteHandler: %s", err))
	}
	return anteHandler
}

func (app *ArchwayApp) PostHandler() sdk.PostHandler {
	postHandler, err := posthandler.NewPostHandler(
		posthandler.HandlerOptions{},
	)
	if err != nil {
		panic(fmt.Errorf("failed to create PostHandler: %s", err))
	}
	return postHandler
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

	// Setup swagger if enabled
	if apiConfig.Swagger {
		if err := RegisterSwaggerAPI(apiSvr); err != nil {
			panic(err)
		}
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
