package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/archway-network/archway/pkg"
	"github.com/archway-network/archway/x/rewards/types"
)

// GetQueryCmd builds query command group for the module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the tracking module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(
		getQueryParamsCmd(),
		getQueryBlockRewardsTrackingCmd(),
		getQueryContractMetadataCmd(),
		getQueryUndistributedPoolFundsCmd(),
	)

	return cmd
}

func getQueryParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query module parameters",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func getQueryBlockRewardsTrackingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block-rewards-tracking",
		Args:  cobra.NoArgs,
		Short: "Query rewards tracking data for the current block height",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.BlockRewardsTracking(cmd.Context(), &types.QueryBlockRewardsTrackingRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func getQueryContractMetadataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contract-metadata [contract-address]",
		Args:  cobra.ExactArgs(1),
		Short: "Query contract metadata (contract rewards parameters)",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			contractAddr, err := pkg.ParseAccAddressArg("contract-address", args[0])
			if err != nil {
				return err
			}

			res, err := queryClient.ContractMetadata(cmd.Context(), &types.QueryContractMetadataRequest{
				ContractAddress: contractAddr.String(),
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Metadata)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func getQueryUndistributedPoolFundsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pool",
		Args:  cobra.NoArgs,
		Short: "Query undistributed rewards pool funds",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.RewardsPool(cmd.Context(), &types.QueryRewardsPoolRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
