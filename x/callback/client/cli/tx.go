package cli

import (
	"github.com/archway-network/archway/pkg"
	"github.com/archway-network/archway/x/callback/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"
)

// GetTxCmd builds tx command group for the module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Transaction commands for the callbacks module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(
		getTxRequestCallbackCmd(),
		getTxCancelCallbackCmd(),
	)

	return cmd
}

func getTxRequestCallbackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request-callback [contract-address] [job-id] [callback-height] [fee-amount]",
		Args:  cobra.ExactArgs(4),
		Short: "Request a new callback for the given contract address and job ID at the given height",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			senderAddr := clientCtx.GetFromAddress()

			contractAddress, err := pkg.ParseAccAddressArg("contract-address", args[0])
			if err != nil {
				return err
			}

			jobID, err := pkg.ParseUint64Arg("job-id", args[1])
			if err != nil {
				return err
			}

			callbackHeight, err := pkg.ParseInt64Arg("callback-height", args[2])
			if err != nil {
				return err
			}

			fees, err := pkg.ParseCoinArg("fee-amount", args[3])
			if err != nil {
				return err
			}

			msg := types.NewMsgRequestCallback(senderAddr, contractAddress, jobID, callbackHeight, fees)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func getTxCancelCallbackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel-callback [contract-address] [job-id] [callback-height]",
		Args:  cobra.ExactArgs(3),
		Short: "Cancel an existing callback given the contract address and its job ID at the specified height",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			senderAddr := clientCtx.GetFromAddress()

			contractAddress, err := pkg.ParseAccAddressArg("contract-address", args[0])
			if err != nil {
				return err
			}

			jobID, err := pkg.ParseUint64Arg("job-id", args[1])
			if err != nil {
				return err
			}

			callbackHeight, err := pkg.ParseInt64Arg("callback-height", args[2])
			if err != nil {
				return err
			}

			msg := types.NewMsgCancelCallback(senderAddr, contractAddress, jobID, callbackHeight)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
