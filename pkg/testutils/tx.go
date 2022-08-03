package testutils

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.FeeTx = MockFeeTx{}

// MockFeeTx is a mock implementation of sdk.FeeTx.
type MockFeeTx struct {
	fees       sdk.Coins
	gas        uint64
	msgs       []sdk.Msg
	feePayer   sdk.AccAddress
	feeGranter sdk.AccAddress
}

type MockFeeTxOption func(tx *MockFeeTx)

// WithMockFeeTxFees option sets the fees of the MockFeeTx.
func WithMockFeeTxFees(fees sdk.Coins) MockFeeTxOption {
	return func(tx *MockFeeTx) {
		tx.fees = fees
	}
}

// WithMockFeeTxMsgs option sets the msgs of the MockFeeTx.
func WithMockFeeTxMsgs(msgs ...sdk.Msg) MockFeeTxOption {
	return func(tx *MockFeeTx) {
		tx.msgs = msgs
	}
}

// WithMockFeeTxPayer option sets the feePayer of the MockFeeTx.
func WithMockFeeTxPayer(payer sdk.AccAddress) MockFeeTxOption {
	return func(tx *MockFeeTx) {
		tx.feePayer = payer
	}
}

// WithMockFeeTxGas option sets the gas limit of the MockFeeTx.
func WithMockFeeTxGas(gas uint64) MockFeeTxOption {
	return func(tx *MockFeeTx) {
		tx.gas = gas
	}
}

// NewMockFeeTx creates a new MockFeeTx instance.
// CONTRACT: tx has no defaults, so it is up to a developer to set options right.
func NewMockFeeTx(opts ...MockFeeTxOption) MockFeeTx {
	tx := MockFeeTx{}
	for _, opt := range opts {
		opt(&tx)
	}

	return tx
}

// GetMsgs implemets the sdk.Tx interface.
func (tx MockFeeTx) GetMsgs() []sdk.Msg {
	return tx.msgs
}

// ValidateBasic implemets the sdk.Tx interface.
func (tx MockFeeTx) ValidateBasic() error {
	return nil
}

// GetGas implements the sdk.FeeTx interface.
func (tx MockFeeTx) GetGas() uint64 {
	return tx.gas
}

// GetFee implements the sdk.FeeTx interface.
func (tx MockFeeTx) GetFee() sdk.Coins {
	return tx.fees
}

// FeePayer implements the sdk.FeeTx interface.
func (tx MockFeeTx) FeePayer() sdk.AccAddress {
	return tx.feePayer
}

// FeeGranter implements the sdk.FeeTx interface.
func (tx MockFeeTx) FeeGranter() sdk.AccAddress {
	return tx.feeGranter
}
