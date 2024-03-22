package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/archway-network/archway/pkg"
	"github.com/archway-network/archway/x/cwerrors/types"
)

// GetTxCmd builds tx command group for the module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Transaction commands for the cwerrors module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(
		getTxSubscribeToErrorCmd(),
	)

	return cmd
}

// getTxSubscribeToErrorCmd returns the command to subscribe to error callbacks for a contract address.
func getTxSubscribeToErrorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "subscribe-to-error [contract-address] [fee-amount]",
		Args:    cobra.ExactArgs(2),
		Short:   "Subscribe to error callbacks for a contract address",
		Aliases: []string{"subscribe"},
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			senderAddr := clientCtx.GetFromAddress()

			fees, err := pkg.ParseCoinArg("fee-amount", args[1])
			if err != nil {
				return err
			}

			msg := types.MsgSubscribeToError{
				Sender:          senderAddr.String(),
				ContractAddress: args[0],
				Fee:             fees,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
