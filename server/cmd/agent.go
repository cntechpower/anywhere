package cmd

import (
	"fmt"

	"github.com/cntechpower/anywhere/server/api/rpc/handler"

	"github.com/spf13/cobra"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "agent admin interface",
	Long:  `agent admin interface.`,
}
var agentListCmd = &cobra.Command{
	Use:   "list",
	Short: "list agents",
	Long:  `list anywhere agents.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := handler.ListAgent(); err != nil {
			fmt.Printf("error query agent list: %v\n", err)
		}
	},
}

func Agent() *cobra.Command {
	agentCmd.AddCommand(agentListCmd)
	return agentCmd
}
