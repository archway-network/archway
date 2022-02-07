package cli

import (
	"encoding/json"
	gstTypes "github.com/archway-network/archway/x/gastracker/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
)

func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        gstTypes.ModuleName,
		Short:                      "Gastracker transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	txCmd.AddCommand(SetContractMetadataCmd())
	return txCmd
}

func SetContractMetadataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-contract-metadata [contract_address] [json_encoded_contract_metadata]",
		Short:   "Set contract metadata",
		Aliases: []string{"set-metadata"},
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			msg, err := parseSetContractMetadataArg(clientCtx.GetFromAddress(), args[0], args[1])
			if err != nil {
				return err
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func parseSetContractMetadataArg(sender sdk.AccAddress, contractAddressStr string, contractMetadataStr string) (gstTypes.MsgSetContractMetadata, error) {
	contractAddress, err := sdk.AccAddressFromBech32(contractAddressStr)
	if err != nil {
		return gstTypes.MsgSetContractMetadata{}, err
	}

	contractMetadata := gstTypes.ContractInstanceMetadata{}
	if err := json.Unmarshal([]byte(contractMetadataStr), &contractMetadata); err != nil {
		return gstTypes.MsgSetContractMetadata{}, err
	}

	return gstTypes.MsgSetContractMetadata{
		Sender:          sender.String(),
		ContractAddress: contractAddress.String(),
		Metadata:        &contractMetadata,
	}, nil
}
