package app

import (
	"encoding/json"
	"os"
	"testing"

	db "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

var emptyWasmOpts []wasmkeeper.Option = nil

func TestArchwaydExport(t *testing.T) {
	db := db.NewMemDB()
	gapp := NewArchwayApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{}, DefaultNodeHome, 0, MakeEncodingConfig(), wasmtypes.EnableAllProposals, EmptyBaseAppOptions{}, emptyWasmOpts)

	genesisState := NewDefaultGenesisState(gapp.AppCodec())
	stateBytes, err := json.MarshalIndent(genesisState, "", "  ")
	require.NoError(t, err)

	// Initialize the chain
	gapp.InitChain(
		abci.RequestInitChain{
			Validators:    []abci.ValidatorUpdate{},
			AppStateBytes: stateBytes,
		},
	)
	gapp.Commit()

	// Making a new app object with the db, so that initchain hasn't been called
	newGapp := NewArchwayApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{}, DefaultNodeHome, 0, MakeEncodingConfig(), wasmtypes.EnableAllProposals, EmptyBaseAppOptions{}, emptyWasmOpts)
	_, err = newGapp.ExportAppStateAndValidators(false, []string{}, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}

// ensure that blocked addresses are properly set in bank keeper
func TestBlockedAddrs(t *testing.T) {
	db := db.NewMemDB()
	gapp := NewArchwayApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{}, DefaultNodeHome, 0, MakeEncodingConfig(), wasmtypes.EnableAllProposals, EmptyBaseAppOptions{}, emptyWasmOpts)

	for acc := range maccPerms {
		t.Run(acc, func(t *testing.T) {
			require.True(t, gapp.Keepers.BankKeeper.BlockedAddr(gapp.Keepers.AccountKeeper.GetModuleAddress(acc)),
				"ensure that blocked addresses are properly set in bank keeper",
			)
		})
	}
}

func TestGetMaccPerms(t *testing.T) {
	dup := GetMaccPerms()
	require.Equal(t, maccPerms, dup, "duplicated module account permissions differed from actual module account permissions")
}

func TestGetEnabledProposals(t *testing.T) {
	cases := map[string]struct {
		proposalsEnabled string
		specificEnabled  string
		expected         []wasmtypes.ProposalType
	}{
		"all disabled": {
			proposalsEnabled: "false",
			expected:         wasmtypes.DisableAllProposals,
		},
		"all enabled": {
			proposalsEnabled: "true",
			expected:         wasmtypes.EnableAllProposals,
		},
		"some enabled": {
			proposalsEnabled: "okay",
			specificEnabled:  "StoreCode,InstantiateContract",
			expected:         []wasmtypes.ProposalType{wasmtypes.ProposalTypeStoreCode, wasmtypes.ProposalTypeInstantiateContract},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ProposalsEnabled = tc.proposalsEnabled
			EnableSpecificProposals = tc.specificEnabled
			proposals := GetEnabledProposals()
			assert.Equal(t, tc.expected, proposals)
		})
	}
}
