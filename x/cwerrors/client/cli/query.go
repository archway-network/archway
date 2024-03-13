package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/archway-network/archway/x/cwerrors/types"
)

// GetQueryCmd builds query command group for the module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the callback module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(
		getQueryErrorsCmd(),
		getQueryIsSubscribedCmd(),
		getQueryParamsCmd(),
	)
	return cmd
}

// getQueryErrorsCmd returns the command to query errors for a contract address.
func getQueryErrorsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "errors [contract_address]",
		Args:  cobra.ExactArgs(1),
		Short: "Query errors for a contract address",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Errors(cmd.Context(), &types.QueryErrorsRequest{
				ContractAddress: args[0],
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

// getQueryIsSubscribedCmd returns the command to query if a contract address is subscribed to errors.
func getQueryIsSubscribedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "is-subscribed [contract_address]",
		Args:  cobra.ExactArgs(1),
		Short: "Query if a contract address is subscribed to errors",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.IsSubscribed(cmd.Context(), &types.QueryIsSubscribedRequest{
				ContractAddress: args[0],
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

// getQueryParamsCmd returns the command to query module parameters.
func getQueryParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query module parameters",
		RunE: func(cmd *cobra.Command, _ []string) error {
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
