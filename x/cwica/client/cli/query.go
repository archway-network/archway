package cli

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/archway-network/archway/x/cwica/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(_ string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 1,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryParams())
	cmd.AddCommand(CmdInterchainAccountCmd())

	return cmd
}

func CmdInterchainAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "interchain-account [owner-address] [connection-id] [interchain-account-id]",
		Short:   "Gets the interchain account address that is associated with the given owner-address, connection-id and interchain-account-id",
		Long:    "Gets the interchain account address that is associated with the given owner-address, connection-id and interchain-account-id. \nThe owner-address is the address of the contract account that is associated with the interchain account. \nThe connection-id is the IBC connection id between the two chains. \nThe interchain-account-id is the identifier of the interchain account that is associated with the owner-address and connection-id.",
		Example: "archway query cwica interchain-account archway14k24jzduc365kywrsvf5ujz4ya6mwymy8vq4q connection-0 1",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.InterchainAccountAddress(cmd.Context(), &types.QueryInterchainAccountAddressRequest{
				OwnerAddress:        args[0],
				ConnectionId:        args[1],
				InterchainAccountId: args[2],
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

func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "fetches the parameters of the cwica module",
		Long:  "fetches the parameters of the cwica module",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
