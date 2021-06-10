package cmd

import (
	"fmt"

	"github.com/cntechpower/anywhere/server/conf"

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
		if err := conf.WriteInitConfigFile(); err != nil {
			fmt.Printf("error reset proxy config: %v\n", err)
		}
	},
}

func Config() *cobra.Command {
	configCmd.AddCommand(resetConfigCmd)
	return configCmd
}
