package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cadence-import",
	Short: "Cadence-import utilities for dealing with imports for local cadence development",
	Long: `Cadence-import uses the flow go sdk to pull deployed contracts
and their dependencies locally, and update their import statments, and updates flow.json
to make it easier to bootstrap new cadence projects`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {

}
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
