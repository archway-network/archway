package testutils

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MockBankKeeper struct {
	GetAllBalancesFn               func(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromAccountToModuleFn func(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccountFn func(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModuleFn  func(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	BlockedAddrFn                  func(addr sdk.AccAddress) bool
}

func (k MockBankKeeper) GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins {
	if k.GetAllBalancesFn == nil {
		panic("not supposed to be called!")
	}
	return k.GetAllBalancesFn(ctx, addr)
}

func (k MockBankKeeper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	if k.SendCoinsFromAccountToModuleFn == nil {
		panic("not supposed to be called!")
	}
	return k.SendCoinsFromAccountToModuleFn(ctx, senderAddr, recipientModule, amt)
}

func (k MockBankKeeper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	if k.SendCoinsFromModuleToAccountFn == nil {
		panic("not supposed to be called!")
	}
	return k.SendCoinsFromModuleToAccountFn(ctx, senderModule, recipientAddr, amt)
}

func (k MockBankKeeper) SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error {
	if k.SendCoinsFromModuleToModuleFn == nil {
		panic("not supposed to be called!")
	}
	return k.SendCoinsFromModuleToModuleFn(ctx, senderModule, recipientModule, amt)
}

func (k MockBankKeeper) BlockedAddr(addr sdk.AccAddress) bool {
	if k.BlockedAddrFn == nil {
		panic("not supposed to be called!")
	}
	return k.BlockedAddrFn(addr)
}
