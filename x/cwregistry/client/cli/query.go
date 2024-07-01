package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/archway-network/archway/x/cwregistry/types"
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

	cmd.AddCommand(CmdQueryCodeMetadata())
	cmd.AddCommand(CmdQueryContractMetadata())
	cmd.AddCommand(CmdQueryCodeSchema())
	cmd.AddCommand(CmdQueryContractSchema())

	return cmd
}

func CmdQueryCodeMetadata() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "code-metadata",
		Short: "",
		Long:  "",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			codeID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			res, err := queryClient.CodeMetadata(context.Background(), &types.QueryCodeMetadataRequest{
				CodeId: codeID,
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

func CmdQueryContractMetadata() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contract-metadata",
		Short: "",
		Long:  "",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			contractAddr := args[0]
			res, err := queryClient.ContractMetadata(context.Background(), &types.QueryContractMetadataRequest{
				ContractAddress: contractAddr,
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

func CmdQueryCodeSchema() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "code-schema",
		Short: "",
		Long:  "",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			codeID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			res, err := queryClient.CodeSchema(context.Background(), &types.QueryCodeSchemaRequest{
				CodeId: codeID,
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

func CmdQueryContractSchema() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contract-schema",
		Short: "",
		Long:  "",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			contractAddr := args[0]
			res, err := queryClient.ContractSchema(context.Background(), &types.QueryContractSchemaRequest{
				ContractAddress: contractAddr,
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
