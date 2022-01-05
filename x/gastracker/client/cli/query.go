package cli

import (
	"context"
	gstTypes "github.com/archway-network/archway/x/gastracker/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
)

func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        gstTypes.ModuleName,
		Short:                      "Querying commands for the gastracker module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	queryCmd.AddCommand(
		GetCmdContractMetadata(),
		GetCmdBlockGasTracking(),
	)
	return queryCmd
}

func GetCmdContractMetadata() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get-contract-metadata [contract-address]",
		Short:   "Get Contract metadata",
		Long:    "Get gastracker module's metadata for the contract",
		Aliases: []string{"get-metadata", "metadata"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			contractAddress, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			queryClient := gstTypes.NewQueryClient(clientCtx)
			resp, err := queryClient.ContractMetadata(context.Background(), &gstTypes.QueryContractMetadataRequest{Address: contractAddress.String()})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetCmdBlockGasTracking() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get-block-gas-tracking",
		Short:   "Get Block Gas Tracking",
		Long:    "Get gastracker module's gas tracking object for current height",
		Aliases: []string{"get-gas-tracking", "get-tracking"},
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := gstTypes.NewQueryClient(clientCtx)
			resp, err := queryClient.BlockGasTracking(context.Background(), &gstTypes.QueryBlockGasTrackingRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
