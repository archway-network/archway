package cli

import "github.com/spf13/cobra"

const (
	flagOwnerAddress   = "owner-address"
	flagRewardsAddress = "rewards-address"
)

func addOwnerAddressFlag(cmd *cobra.Command) {
	cmd.Flags().String(flagOwnerAddress, "", "Address of the contract owner (bech 32)")
}

func addRewardsAddressFlag(cmd *cobra.Command) {
	cmd.Flags().String(flagRewardsAddress, "", "Rewards address to distribute contract rewards to (bech 32)")
}
