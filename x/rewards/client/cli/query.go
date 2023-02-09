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
		Short:                      "Querying commands for the rewards module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(
		getQueryParamsCmd(),
		getQueryBlockRewardsTrackingCmd(),
		getQueryContractMetadataCmd(),
		getQueryUndistributedPoolFundsCmd(),
		getQueryEstimateTxFeesCmd(),
		getQueryOutstandingRewardsCmd(),
		getQueryRewardsRecordsCmd(),
		getQueryContractFlatFeeCmd(),
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
		Short: "Query the undistributed rewards pool (ready for withdrawal) and the treasury pool funds",
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

func getQueryEstimateTxFeesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "estimate-fees [gas-limit] [contract-address]",
		Args:  cobra.MinimumNArgs(1),
		Short: "Query transaction fees estimation for a give gas limit, optionally takes in contract address to include the flat fees in the estimate",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			gasLimit, err := pkg.ParseUint64Arg("gas-limit", args[0])
			if err != nil {
				return err
			}

			req := types.QueryEstimateTxFeesRequest{
				GasLimit: gasLimit,
			}

			if len(args) > 1 {
				contractAddr, err := pkg.ParseAccAddressArg("contract-address", args[1])
				if err != nil {
					return err
				}
				req.ContractAddress = contractAddr.String()
			}

			res, err := queryClient.EstimateTxFees(cmd.Context(), &req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func getQueryOutstandingRewardsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "outstanding-rewards [rewards-address]",
		Args:  cobra.ExactArgs(1),
		Short: "Query current credited rewards for a given address (the address set in contract(s) metadata rewards_address field)",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			rewardsAddr, err := pkg.ParseAccAddressArg("rewards-address", args[0])
			if err != nil {
				return err
			}

			res, err := queryClient.OutstandingRewards(cmd.Context(), &types.QueryOutstandingRewardsRequest{
				RewardsAddress: rewardsAddr.String(),
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func getQueryRewardsRecordsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rewards-records [rewards-address]",
		Args:  cobra.ExactArgs(1),
		Short: "Query rewards records stored for a given address (the address set in contract(s) metadata rewards_address field) with pagination",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			rewardsAddr, err := pkg.ParseAccAddressArg("rewards-address", args[0])
			if err != nil {
				return err
			}

			pageReq, err := pkg.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.RewardsRecords(cmd.Context(), &types.QueryRewardsRecordsRequest{
				RewardsAddress: rewardsAddr.String(),
				Pagination:     pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "rewards-records")

	return cmd
}

func getQueryContractFlatFeeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "flat-fee [contract-address]",
		Args:  cobra.ExactArgs(1),
		Short: "Query contract flat-fee",
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

			res, err := queryClient.FlatFee(cmd.Context(), &types.QueryFlatFeeRequest{
				ContractAddress: contractAddr.String(),
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.FlatFeeAmount)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
