package cmd

import (
	"fmt"

	"github.com/jrkhan/cadence-import/pkg/imports"
	"github.com/spf13/cobra"
)

var importer = imports.Importer{}

var getCmd = &cobra.Command{
	Use:        "get [contractName] [flags] ",
	Short:      "make a copy of a deployed smart contract and recursively retrieve dependencies",
	Long:       `Retreives already deployed contracts and dependencies to bootrap local development`,
	Args:       cobra.MinimumNArgs(1),
	ArgAliases: []string{"contractName"},
	RunE: func(cmd *cobra.Command, args []string) error {
		contractName := args[0]
		defer func() {
			if importer.Verbose {
				return
			}
			e := recover()
			var err error
			if e != nil {
				err = fmt.Errorf("🛑  %v", e)
				fmt.Print(err)
			}
		}()
		importer.Get(imports.ReaderWriter{}, contractName)
		return nil
	},
}

func init() {
	getCmd.Flags().StringVarP(&importer.Network, "network", "n", "mainnet", "network to use to retrieve contracts")
	getCmd.Flags().StringVarP(&importer.Address, "address", "a", "", "the address of the contract")
	getCmd.Flags().BoolVar(&importer.Verbose, "verbose", false, "display detailed error messaging")

	rootCmd.AddCommand(getCmd)
}
