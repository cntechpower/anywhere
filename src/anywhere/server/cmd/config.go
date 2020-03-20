package cmd

import (
	"anywhere/server/anywhereServer"
	"fmt"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "config file admin interface",
	Long:  `config file admin interface.`,
}
var resetConfigCmd = &cobra.Command{
	Use:   "reset",
	Short: "reset system config file",
	Long:  `reset system config file 'anywhered.json'`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := anywhereServer.WriteInitConfigFile(); err != nil {
			fmt.Printf("error reset proxy config: %v\n", err)
		}
	},
}

func GetConfigCmd() *cobra.Command {
	configCmd.AddCommand(resetConfigCmd)
	return configCmd
}
