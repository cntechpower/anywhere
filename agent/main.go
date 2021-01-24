package main

import (
	"context"
	"fmt"

	"github.com/cntechpower/anywhere/agent/anywhereAgent"
	"github.com/cntechpower/anywhere/agent/handler"
	"github.com/cntechpower/anywhere/util"
	"github.com/cntechpower/utils/log"

	"github.com/spf13/cobra"
)

var serverPort int
var serverIp, agentId, user, password, certFile, keyFile, caFile string
var agentGroup, version string

var grpcAddress string
var connIdToKill int

func main() {
	log.InitLogger("")
	var rootCmd = &cobra.Command{
		Use:   "anywhere --help",
		Short: "This is A Proxy Agent ",
		Long:  "anywhere agent - " + version,
		Run: func(cmd *cobra.Command, args []string) {
			if err := run(cmd, args); err != nil {
				panic(err)
			}
		},
	}
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
			if err := handler.ListConns(grpcAddress); err != nil {
				fmt.Printf("error query conn list: %v\n", err)
			}
		},
	}

	var connKillCmd = &cobra.Command{
		Use:   "kill",
		Short: "kill conn",
		Long:  `kill anywhere conn.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := handler.KillConn(grpcAddress, connIdToKill); err != nil {
				fmt.Printf("error query agent list: %v\n", err)
			}
		},
	}
	var connFlushCmd = &cobra.Command{
		Use:   "flush",
		Short: "flush conn",
		Long:  `flush anywhere conn.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := handler.FlushConns(grpcAddress); err != nil {
				fmt.Printf("error query agent list: %v\n", err)
			}
		},
	}
	var statusCmd = &cobra.Command{
		Use:   "status",
		Short: "status",
		Long:  `show agent status`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := handler.ShowStatus(grpcAddress); err != nil {
				fmt.Printf("error query agent status: %v\n", err)
			}
		},
	}
	rootCmd.PersistentFlags().StringVarP(&serverIp, "server-ip", "s", "127.0.0.1", "anywhered server address")
	rootCmd.PersistentFlags().IntVarP(&serverPort, "server-port", "p", 1111, "anywhered server port")
	rootCmd.PersistentFlags().StringVarP(&agentGroup, "agent-group", "g", "asia-shanghai", "anywhere agent group")
	rootCmd.PersistentFlags().StringVarP(&agentId, "agent-id", "i", "anywhere-agent-1", "anywhere agent id")
	rootCmd.PersistentFlags().StringVar(&grpcAddress, "grpc-address", "127.0.0.1:1110", "anywhere agent grpc address")
	rootCmd.PersistentFlags().StringVarP(&user, "user", "u", "none", "anywhere user")
	rootCmd.PersistentFlags().StringVarP(&password, "pass", "", "none", "anywhere password")
	rootCmd.PersistentFlags().StringVar(&certFile, "cert", "credential/client.crt", "cert file")
	rootCmd.PersistentFlags().StringVar(&keyFile, "key", "credential/client.key", "key file")
	rootCmd.PersistentFlags().StringVar(&caFile, "ca", "credential/ca.crt", "ca file")
	connCmd.AddCommand(connListCmd)
	connKillCmd.PersistentFlags().IntVar(&connIdToKill, "id", -1, "conn id to kill")
	connCmd.AddCommand(connKillCmd)
	connCmd.AddCommand(connFlushCmd)
	rootCmd.AddCommand(connCmd)
	rootCmd.AddCommand(statusCmd)
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}

}

func run(_ *cobra.Command, _ []string) error {
	h := log.NewHeader("agentMain")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	a := anywhereAgent.InitAnyWhereAgent(agentGroup, agentId, serverIp, user, password, serverPort)
	if err := a.SetCredentials(certFile, keyFile, caFile); err != nil {
		return err
	}
	go a.Start(ctx)

	go util.ListenTTINSignalLoop()
	serverExitChan := util.ListenKillSignal()
	rpcExitChan := make(chan error, 0)
	go handler.StartRpcServer(a, grpcAddress, rpcExitChan)

	select {
	case err := <-rpcExitChan:
		h.Fatalf("Grpc existing unexpected: %v", err)
	case <-serverExitChan:
		h.Infof("Agent Existing")
	}

	return nil
}
