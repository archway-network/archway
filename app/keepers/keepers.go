package keepers

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	nftkeeper "github.com/cosmos/cosmos-sdk/x/nft/keeper"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"

	icahostkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/keeper"
	ibcfeekeeper "github.com/cosmos/ibc-go/v7/modules/apps/29-fee/keeper"
	ibctransferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"

	callbackKeeper "github.com/archway-network/archway/x/callback/keeper"
	rewardsKeeper "github.com/archway-network/archway/x/rewards/keeper"
	trackingKeeper "github.com/archway-network/archway/x/tracking/keeper"
)

type ArchwayKeepers struct {
	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.Keeper
	AuthzKeeper           authzkeeper.Keeper
	CapabilityKeeper      *capabilitykeeper.Keeper
	ConsensusParamsKeeper consensusparamkeeper.Keeper
	CrisisKeeper          crisiskeeper.Keeper
	FeeGrantKeeper        feegrantkeeper.Keeper
	GroupKeeper           groupkeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	MintKeeper            mintkeeper.Keeper
	NFTKeeper             nftkeeper.Keeper
	GovKeeper             govkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	SlashingKeeper        slashingkeeper.Keeper
	ParamsKeeper          paramskeeper.Keeper
	EvidenceKeeper        evidencekeeper.Keeper
	UpgradeKeeper         upgradekeeper.Keeper

	// IBC Keepers
	IBCKeeper      *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	IBCFeeKeeper   ibcfeekeeper.Keeper
	TransferKeeper ibctransferkeeper.Keeper
	ICAHostKeeper  icahostkeeper.Keeper

	// Wasm Keepers
	WASMKeeper wasmkeeper.Keeper

	// Archway Keepers
	TrackingKeeper trackingKeeper.Keeper
	RewardsKeeper  rewardsKeeper.Keeper
	CallbackKeeper callbackKeeper.Keeper
}

type ArchwayScopedKeepers struct {
	ScopedIBCKeeper      capabilitykeeper.ScopedKeeper
	ScopedICAHostKeeper  capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper capabilitykeeper.ScopedKeeper
	ScopedIBCFeeKeeper   capabilitykeeper.ScopedKeeper
	ScopedWASMKeeper     capabilitykeeper.ScopedKeeper
}
