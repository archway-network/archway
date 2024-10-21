package keepers

import (
	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	nftkeeper "cosmossdk.io/x/nft/keeper"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"

	"github.com/archway-network/archway/x/cwfees"

	ibchookskeeper "github.com/cosmos/ibc-apps/modules/ibc-hooks/v8/keeper"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/keeper"
	icahostkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/keeper"
	ibcfeekeeper "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/keeper"
	ibctransferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"

	callbackKeeper "github.com/archway-network/archway/x/callback/keeper"
	cwerrorsKeeper "github.com/archway-network/archway/x/cwerrors/keeper"
	cwicaKeeper "github.com/archway-network/archway/x/cwica/keeper"
	oracleKeeper "github.com/archway-network/archway/x/oracle/keeper"
	rewardsKeeper "github.com/archway-network/archway/x/rewards/keeper"
	trackingKeeper "github.com/archway-network/archway/x/tracking/keeper"
)

type ArchwayKeepers struct {
	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.Keeper
	CapabilityKeeper      *capabilitykeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	SlashingKeeper        slashingkeeper.Keeper
	MintKeeper            mintkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	GovKeeper             govkeeper.Keeper
	GroupKeeper           groupkeeper.Keeper
	NFTKeeper             nftkeeper.Keeper
	CrisisKeeper          crisiskeeper.Keeper
	UpgradeKeeper         upgradekeeper.Keeper
	ParamsKeeper          paramskeeper.Keeper
	ConsensusParamsKeeper consensusparamkeeper.Keeper
	IBCKeeper             *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	IBCFeeKeeper          ibcfeekeeper.Keeper
	IBCHooksKeeper        ibchookskeeper.Keeper
	ICAControllerKeeper   icacontrollerkeeper.Keeper
	ICAHostKeeper         icahostkeeper.Keeper
	EvidenceKeeper        evidencekeeper.Keeper
	TransferKeeper        ibctransferkeeper.Keeper
	FeeGrantKeeper        feegrantkeeper.Keeper
	AuthzKeeper           authzkeeper.Keeper
	WASMKeeper            wasmkeeper.Keeper

	// Archway Keepers
	TrackingKeeper trackingKeeper.Keeper
	RewardsKeeper  rewardsKeeper.Keeper
	CWFeesKeeper   cwfees.Keeper
	CallbackKeeper callbackKeeper.Keeper
	CWErrorsKeeper cwerrorsKeeper.Keeper
	CWICAKeeper    cwicaKeeper.Keeper
	OracleKeeper   oracleKeeper.Keeper
}
