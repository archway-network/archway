// DONTCOVER
package app

import (
	errorsmod "cosmossdk.io/errors"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	ibcante "github.com/cosmos/ibc-go/v7/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"

	rewardsAnte "github.com/archway-network/archway/x/rewards/ante"
	rewardsKeeper "github.com/archway-network/archway/x/rewards/keeper"
	trackingAnte "github.com/archway-network/archway/x/tracking/ante"
	trackingKeeper "github.com/archway-network/archway/x/tracking/keeper"
)

// HandlerOptions extend the SDK's AnteHandler options by requiring the IBC
// channel keeper.
type HandlerOptions struct {
	ante.HandlerOptions

	IBCKeeper             *ibckeeper.Keeper
	WasmConfig            *wasmTypes.WasmConfig
	RewardsAnteBankKeeper rewardsAnte.BankKeeper

	TXCounterStoreKey storetypes.StoreKey

	TrackingKeeper trackingKeeper.Keeper
	RewardsKeeper  rewardsKeeper.Keeper

	Codec codec.BinaryCodec
}

func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "account keeper is required for AnteHandler")
	}
	if options.BankKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "bank keeper is required for AnteHandler")
	}
	if options.SignModeHandler == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}
	if options.WasmConfig == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "wasm config is required for ante builder")
	}
	if options.TXCounterStoreKey == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "tx counter key is required for ante builder")
	}

	if options.IBCKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "ibc keeper is required for ante builder")
	}

	sigGasConsumer := options.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = ante.DefaultSigVerificationGasConsumer
	}

	anteDecorators := []sdk.AnteDecorator{
		// Outermost AnteDecorator (SetUpContext must be called first)
		ante.NewSetUpContextDecorator(),
		// After setup context to enforce limits early
		wasmkeeper.NewLimitSimulationGasDecorator(options.WasmConfig.SimulationGasLimit),
		wasmkeeper.NewCountTXDecorator(options.TXCounterStoreKey),
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		// Custom Archway minimum fee checker
		rewardsAnte.NewMinFeeDecorator(options.Codec, options.RewardsKeeper),
		// Custom Archway interceptor to track new transactions
		trackingAnte.NewTxGasTrackingDecorator(options.TrackingKeeper),
		// Custom Archway fee deduction, which splits fees between x/rewards and x/auth fee collector
		rewardsAnte.NewDeductFeeDecorator(options.Codec, options.AccountKeeper, options.RewardsAnteBankKeeper, options.FeegrantKeeper, options.RewardsKeeper),
		// SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewSetPubKeyDecorator(options.AccountKeeper),
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, sigGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}
