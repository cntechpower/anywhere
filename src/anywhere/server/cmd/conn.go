package cmd

import (
	"anywhere/server/handler/rpcHandler"
	"fmt"

	"github.com/spf13/cobra"
)

var agentId string
var connIdToKill int
var userName string

var connCmd = &cobra.Command{
	Use:   "conn",
	Short: "conn admin interface",
	Long:  `conn admin interface.`,
}

var connListCmd = &cobra.Command{
	Use:   "list",
	Short: "list conns",
	Long:  `list anywhere conns.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := rpcHandler.ListConns(agentId); err != nil {
			fmt.Printf("error query conn list: %v\n", err)
		}
	},
}

var connKillCmd = &cobra.Command{
	Use:   "kill",
	Short: "kill conn",
	Long:  `kill anywhere conn.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := rpcHandler.KillConn(userName, agentId, connIdToKill); err != nil {
			fmt.Printf("error query agent list: %v\n", err)
		}
	},
}
var connFlushCmd = &cobra.Command{
	Use:   "flush",
	Short: "flush conn",
	Long:  `flush anywhere conn.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := rpcHandler.FlushConns(); err != nil {
			fmt.Printf("error query agent list: %v\n", err)
		}
	},
}

func GetConnCmd() *cobra.Command {
	connCmd.PersistentFlags().StringVar(&userName, "user", "", "user name ")
	connListCmd.PersistentFlags().StringVar(&agentId, "agent-id", "", "agent id to list, leave blank to list all agent")
	connCmd.AddCommand(connListCmd)
	connKillCmd.PersistentFlags().StringVar(&agentId, "agent-id", "anywhere-agent-1", "agent id to delete conn")
	connKillCmd.PersistentFlags().IntVar(&connIdToKill, "conn-id", -1, "conn id to kill")
	connCmd.AddCommand(connKillCmd)
	connCmd.AddCommand(connFlushCmd)
	return connCmd
}
