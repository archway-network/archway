package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/archway-network/archway/pkg"
	"github.com/archway-network/archway/x/rewards/types"
)

// GetTxCmd builds tx command group for the module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Transaction commands for the rewards module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(
		getTxSetContractMetadataCmd(),
		getTxWithdrawRewardsCmd(),
	)

	return cmd
}

func getTxSetContractMetadataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-contract-metadata [contract-address]",
		Args:  cobra.ExactArgs(1),
		Short: "Create / modify contract metadata (contract rewards parameters)",
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

			ownerAddress, err := pkg.ParseAccAddressFlag(cmd, flagOwnerAddress, false)
			if err != nil {
				return err
			}

			rewardsAddress, err := pkg.ParseAccAddressFlag(cmd, flagRewardsAddress, false)
			if err != nil {
				return err
			}

			msg := types.NewMsgSetContractMetadata(senderAddr, contractAddress, ownerAddress, rewardsAddress)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	addOwnerAddressFlag(cmd)
	addRewardsAddressFlag(cmd)

	return cmd
}

func getTxWithdrawRewardsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-rewards",
		Args:  cobra.NoArgs,
		Short: "Withdraw all current credited rewards for a given rewards address",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			senderAddr := clientCtx.GetFromAddress()

			msg := types.NewMsgWithdrawRewards(senderAddr)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
