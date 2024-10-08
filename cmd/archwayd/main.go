package main

import (
	"fmt"
	"os"
	"runtime/debug"

	cosmwasm "github.com/CosmWasm/wasmvm"
	"github.com/spf13/cobra"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"github.com/archway-network/archway/app"
)

const ArchwayASCII = `
 #####   #####     #####   ##   ##  ##   ##   #####    ##  ##
#######  #######  #######  ##   ##  ##   ##  #######   ##  ##
##   ##       ##  ##   ##  ##   ##  ## # ##  ##   ##   ##  ##
##   ##  ######   ##       #### ##  #######  ##   ##    ####
#######  ##   ##  ##   ##  ##   ##  #######  #######     ##
##   ##  ##   ##  #######  ##   ##  ### ###  ##   ##     ##
##   ##  ##   ##   #####   ##   ##  ##   ##  ##   ##     ##

`

func main() {
	rootCmd, _ := NewRootCmd()

	rootCmd.AddCommand(ensureLibWasmVM())

	if err := svrcmd.Execute(rootCmd, "ARCHWAY", app.DefaultNodeHome); err != nil {
		_, _ = fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}

func ensureLibWasmVM() *cobra.Command {
	return &cobra.Command{
		Use:   "ensure-binary",
		Short: "ensures the binary is correctly built",
		RunE: func(cmd *cobra.Command, args []string) error {
			got, err := cosmwasm.LibwasmvmVersion()
			if err != nil {
				return fmt.Errorf("unable to detect the present libwasmvm version: %w", err)
			}

			expected, err := getExpectedLibwasmVersion()
			if err != nil {
				return fmt.Errorf("unable to detect the expected libwasmvm version: %w", err)
			}

			expected = expected[1:]

			if got != expected {
				return fmt.Errorf("libwasmvm version mismatch, wanted: %s, got: %s", expected, got)
			}

			cmd.Println("OK")
			return nil
		},
	}
}

func getExpectedLibwasmVersion() (string, error) {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return "", fmt.Errorf("can't read build info")
	}
	for _, d := range buildInfo.Deps {
		if d.Path != "github.com/CosmWasm/wasmvm" {
			continue
		}
		if d.Replace != nil {
			return d.Replace.Version, nil
		}
		return d.Version, nil
	}
	return "", fmt.Errorf("unable to detect the expected libwasmvm version")
}
