package cmd

import (
	"fmt"

	"github.com/jrkhan/cadence-import/pkg/imports"
	"github.com/spf13/cobra"
)

var Network string
var getCmd = &cobra.Command{
	Use:        "get [contractName] [flags] ",
	Short:      "make a copy of a deployed smart contract and recursively retrieve dependencies",
	Long:       `Retreives already deployed contracts and dependencies to bootrap local development`,
	Args:       cobra.MinimumNArgs(1),
	ArgAliases: []string{"contractName"},
	RunE: func(cmd *cobra.Command, args []string) error {
		defer func() {
			e := recover()
			var err error
			if e != nil {
				err = fmt.Errorf("🛑  %v", e)
			}
			fmt.Print(err)
		}()
		imports.GetImport(imports.ReaderWriter{}, Network, args[0])
		return nil
	},
}

func init() {
	getCmd.Flags().StringVarP(&Network, "network", "n", "mainnet", "network to use to retrieve contracts")
	rootCmd.AddCommand(getCmd)
}
