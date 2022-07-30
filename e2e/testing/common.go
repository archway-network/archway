package e2eTesting

import (
	wasmKeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// GetStringEventAttribute returns TX response event attribute string value by type and attribute key.
func GetStringEventAttribute(events []abci.Event, eventType, attrKey string) string {
	for _, event := range events {
		if event.Type != eventType {
			continue
		}

		for _, attr := range event.Attributes {
			if string(attr.Key) != attrKey {
				continue
			}

			return string(attr.Value)
		}
	}

	return ""
}

// GenAccounts generates a list of accounts and private keys for them.
func GenAccounts(num uint) ([]sdk.AccAddress, []cryptotypes.PrivKey) {
	addrs := make([]sdk.AccAddress, 0, num)
	privKeys := make([]cryptotypes.PrivKey, 0, num)

	for i := 0; i < cap(addrs); i++ {
		privKey := secp256k1.GenPrivKey()

		addrs = append(addrs, sdk.AccAddress(privKey.PubKey().Address()))
		privKeys = append(privKeys, privKey)
	}

	return addrs, privKeys
}

// GenContractAddresses generates a list of contract addresses (codeID and instanceID are sequential).
func GenContractAddresses(num uint) []sdk.AccAddress {
	addrs := make([]sdk.AccAddress, 0, num)

	for i := 0; i < cap(addrs); i++ {
		contractAddr := wasmKeeper.BuildContractAddress(uint64(i), uint64(i))
		addrs = append(addrs, contractAddr)
	}

	return addrs
}
