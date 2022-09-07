package cli

import "github.com/spf13/cobra"

const (
	flagOwnerAddress   = "owner-address"
	flagRewardsAddress = "rewards-address"
	flagRecordsLimit   = "records-limit"
	flagRecordIDs      = "record-ids"
)

func addOwnerAddressFlag(cmd *cobra.Command) {
	cmd.Flags().String(flagOwnerAddress, "", "Address of the contract owner (bech 32)")
}

func addRewardsAddressFlag(cmd *cobra.Command) {
	cmd.Flags().String(flagRewardsAddress, "", "Rewards address to distribute contract rewards to (bech 32)")
}

func addRecordsLimitFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64(flagRecordsLimit, 0, "Max number of rewards records to use (value can not be higher than the MaxWithdrawRecords module param")
}

func addRecordIDsFlag(cmd *cobra.Command) {
	cmd.Flags().StringSlice(flagRecordIDs, []string{}, "Rewards record IDs to use (number of IDs can not be higher than the MaxWithdrawRecords module param")
}
