package cmd

import (
	"fmt"

	"github.com/cntechpower/anywhere/server/api/rpc/handler"

	"github.com/spf13/cobra"
)

var connIdToKill int

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
		if err := handler.ListConns(""); err != nil {
			fmt.Printf("error query conn list: %v\n", err)
		}
	},
}

var connKillCmd = &cobra.Command{
	Use:   "kill",
	Short: "kill conn",
	Long:  `kill anywhere conn.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := handler.KillConn(int64(connIdToKill)); err != nil {
			fmt.Printf("error query agent list: %v\n", err)
		}
	},
}
var connFlushCmd = &cobra.Command{
	Use:   "flush",
	Short: "flush conn",
	Long:  `flush anywhere conn.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := handler.FlushConns(); err != nil {
			fmt.Printf("error query agent list: %v\n", err)
		}
	},
}

func Conn() *cobra.Command {
	connCmd.AddCommand(connListCmd)
	connKillCmd.PersistentFlags().IntVar(&connIdToKill, "conn-id", -1, "conn id to kill")
	connCmd.AddCommand(connKillCmd)
	connCmd.AddCommand(connFlushCmd)
	return connCmd
}
