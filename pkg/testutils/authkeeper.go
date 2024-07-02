package testutils

import (
	"context"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MockAuthKeeper struct {
	GetModuleAccountFn func(ctx context.Context, name string) sdk.ModuleAccountI
	GetAccountFn       func(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
}

func (k MockAuthKeeper) GetModuleAccount(ctx context.Context, name string) sdk.ModuleAccountI {
	if k.GetModuleAccountFn == nil {
		panic("not supposed to be called!")
	}
	return k.GetModuleAccountFn(ctx, name)
}

func (k MockAuthKeeper) GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI {
	if k.GetAccountFn == nil {
		panic("not supposed to be called!")
	}
	return k.GetAccountFn(ctx, addr)
}

type MockModuleAccount struct {
	Address string
}

func (MockModuleAccount) GetName() string           { return "" }
func (MockModuleAccount) GetPermissions() []string  { return nil }
func (MockModuleAccount) HasPermission(string) bool { return true }
func (a MockModuleAccount) GetAddress() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(a.Address)
}
func (MockModuleAccount) SetAddress(sdk.AccAddress) error    { return nil }
func (MockModuleAccount) GetPubKey() cryptotypes.PubKey      { return nil }
func (MockModuleAccount) SetPubKey(cryptotypes.PubKey) error { return nil }
func (MockModuleAccount) GetAccountNumber() uint64           { return 0 }
func (MockModuleAccount) SetAccountNumber(uint64) error      { return nil }
func (MockModuleAccount) GetSequence() uint64                { return 0 }
func (MockModuleAccount) SetSequence(uint64) error           { return nil }
func (MockModuleAccount) String() string                     { return "" }
func (MockModuleAccount) Reset()                             {}
func (MockModuleAccount) ProtoMessage()                      {}
