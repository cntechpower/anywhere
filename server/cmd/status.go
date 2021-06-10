package cmd

import (
	"fmt"

	"github.com/cntechpower/anywhere/server/rpc/handler"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "status interface",
	Long:  `status interface.`,
}

var emailCmd = &cobra.Command{
	Use:   "email",
	Short: "report server status to email",
	Long:  `report server status to email.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := handler.SendReport(); err != nil {
			fmt.Printf("send report error: %v\n", err)
		}
	},
}

func Status() *cobra.Command {
	statusCmd.AddCommand(emailCmd)
	return statusCmd
}
