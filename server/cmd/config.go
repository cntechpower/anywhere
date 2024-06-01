package cmd

import (
	"fmt"

	"github.com/cntechpower/anywhere/dao/config"

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
			fmt.Printf("error reset proxy config: %+v\n", err)
		}
	},
}

var migrateConfigCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrate system config file to db",
	Long:  `migrate config file 'anywhered.json' to db`,
	Run: func(cmd *cobra.Command, args []string) {
		err := config.Migrate()
		if err != nil {
			fmt.Printf("error save new proxy config: %+v\n", err)
		}
	},
}

func Config() *cobra.Command {
	configCmd.AddCommand(resetConfigCmd)
	configCmd.AddCommand(migrateConfigCmd)
	return configCmd
}
