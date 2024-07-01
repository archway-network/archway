package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/archway-network/archway/pkg"
	"github.com/archway-network/archway/x/cwregistry/types"
)

// GetTxCmd builds tx command group for the module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Transaction commands for the cwregistry module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdRegisterCode())
	cmd.AddCommand(CmdRegisterContract())

	return cmd
}

func CmdRegisterContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-contract [contract-address]",
		Args:  cobra.ExactArgs(1),
		Short: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			senderAddr := clientCtx.GetFromAddress()
			contractAddr := args[0]

			sourceMetadata := parseSourceMetadata(cmd)
			sourceBuilder := parseSourceBuilder(cmd)
			schema := parseSchema(cmd)
			contacts, err := pkg.GetStringSliceFlag(cmd, flagContacts, true)
			if err != nil {
				return err
			}
			msg := types.MsgRegisterContract{
				Sender:          senderAddr.String(),
				ContractAddress: contractAddr,
				SourceMetadata:  &sourceMetadata,
				SourceBuilder:   &sourceBuilder,
				Schema:          schema,
				Contacts:        contacts,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	addFlags(cmd)

	return cmd
}

func CmdRegisterCode() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-code [code-id]",
		Args:  cobra.ExactArgs(1),
		Short: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			senderAddr := clientCtx.GetFromAddress()
			codeID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			sourceMetadata := parseSourceMetadata(cmd)
			sourceBuilder := parseSourceBuilder(cmd)
			schema := parseSchema(cmd)
			contacts, err := pkg.GetStringSliceFlag(cmd, flagContacts, true)
			if err != nil {
				return err
			}
			msg := types.MsgRegisterCode{
				Sender:         senderAddr.String(),
				CodeId:         codeID,
				SourceMetadata: &sourceMetadata,
				SourceBuilder:  &sourceBuilder,
				Schema:         schema,
				Contacts:       contacts,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	addFlags(cmd)

	return cmd
}
