package cli

import (
	"os"

	"github.com/archway-network/archway/x/cwregistry/types"
	"github.com/spf13/cobra"
)

const (
	// Source Metadata
	flagSourceRepository = "source-repository"
	flagSourceTag        = "source-tag"
	flagSourceLicense    = "source-license"
	// Source Builder
	flagSourceImage        = "source-image"
	flagSourceImageTag     = "source-image-tag"
	flagSourceContractName = "source-contract-name"

	flagContacts   = "contacts"
	flagSchemaPath = "schema-path"
)

func addFlags(cmd *cobra.Command) {
	addContactsFlag(cmd)
	addSchemaPathFlag(cmd)
	addSourceRepositoryFlag(cmd)
	addSourceTagFlag(cmd)
	addSourceLicenseFlag(cmd)
	addSourceImageFlag(cmd)
	addSourceImageTagFlag(cmd)
	addSourceContractNameFlag(cmd)
}

func addSourceRepositoryFlag(cmd *cobra.Command) {
	cmd.Flags().String(flagSourceRepository, "", "The link to the code repository. e.g https://github.com/archway-network/archway")
}

func addSourceTagFlag(cmd *cobra.Command) {
	cmd.Flags().String(flagSourceTag, "", "The tag of the code repository. e.g v0.1.0")
}

func addSourceLicenseFlag(cmd *cobra.Command) {
	cmd.Flags().String(flagSourceLicense, "", "The license of the code repository. e.g Apache-2.0")
}

func addSourceImageFlag(cmd *cobra.Command) {
	cmd.Flags().String(flagSourceImage, "", "The docker image of the contract. e.g cosmwasm/rust-optimizer")
}

func addSourceImageTagFlag(cmd *cobra.Command) {
	cmd.Flags().String(flagSourceImageTag, "", "The tag of the docker image. e.g 0.12.6")
}

func addSourceContractNameFlag(cmd *cobra.Command) {
	cmd.Flags().String(flagSourceContractName, "", "The name of the contract in the docker image. e.g counter.wasm")
}

func addContactsFlag(cmd *cobra.Command) {
	cmd.Flags().StringSlice(flagContacts, []string{}, "The list of contacts for the contract. e.g admin@dapp.com,security@dapp.com")
}

func addSchemaPathFlag(cmd *cobra.Command) {
	cmd.Flags().String(flagSchemaPath, "", "The path to the schema file. e.g ./schema.json")
}

func parseSourceMetadata(cmd *cobra.Command) types.SourceMetadata {
	repository, _ := cmd.Flags().GetString(flagSourceRepository)
	tag, _ := cmd.Flags().GetString(flagSourceTag)
	license, _ := cmd.Flags().GetString(flagSourceLicense)
	sourceMetadata := types.SourceMetadata{
		Repository: repository,
		Tag:        tag,
		License:    license,
	}
	return sourceMetadata
}

func parseSourceBuilder(cmd *cobra.Command) types.SourceBuilder {
	image, _ := cmd.Flags().GetString(flagSourceImage)
	imageTag, _ := cmd.Flags().GetString(flagSourceImageTag)
	contractName, _ := cmd.Flags().GetString(flagSourceContractName)
	sourceBuilder := types.SourceBuilder{
		Image:        image,
		Tag:          imageTag,
		ContractName: contractName,
	}
	return sourceBuilder
}

func parseSchema(cmd *cobra.Command) string {
	schemaPath, _ := cmd.Flags().GetString(flagSchemaPath)
	fileContent, err := os.ReadFile(schemaPath)
	if err != nil {
		return ""
	}
	return string(fileContent)
}
