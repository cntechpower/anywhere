package main

import (
	"anywhere/agent/anywhereAgent"
	"anywhere/agent/handler"
	"anywhere/log"
	"anywhere/util"
	"fmt"

	"github.com/spf13/cobra"
)

var serverPort int
var serverIp, agentId, certFile, keyFile, caFile string
var version string

var grpcAddress string
var connIdToKill int

func main() {
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
	rootCmd.PersistentFlags().StringVarP(&serverIp, "server-ip", "s", "127.0.0.1", "anywhered server address")
	rootCmd.PersistentFlags().IntVarP(&serverPort, "server-port", "p", 1111, "anywhered server port")
	rootCmd.PersistentFlags().StringVarP(&agentId, "agent-id", "i", "anywhere-agent-1", "anywhere agent id")
	rootCmd.PersistentFlags().StringVarP(&grpcAddress, "grpc-address", "g", "127.0.0.1:1110", "anywhere agent grpc address")
	rootCmd.PersistentFlags().StringVar(&certFile, "cert", "credential/client.crt", "cert file")
	rootCmd.PersistentFlags().StringVar(&keyFile, "key", "credential/client.key", "key file")
	rootCmd.PersistentFlags().StringVar(&caFile, "ca", "credential/ca.crt", "ca file")
	connCmd.AddCommand(connListCmd)
	connKillCmd.PersistentFlags().IntVar(&connIdToKill, "id", -1, "conn id to kill")
	connCmd.AddCommand(connKillCmd)
	connCmd.AddCommand(connFlushCmd)
	rootCmd.AddCommand(connCmd)
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}

}

func run(_ *cobra.Command, _ []string) error {

	log.InitLogger("")
	a := anywhereAgent.InitAnyWhereAgent(agentId, serverIp, serverPort)
	if err := a.SetCredentials(certFile, keyFile, caFile); err != nil {
		return err
	}
	go a.Start()

	go util.ListenTTINSignalLoop()
	serverExitChan := util.ListenKillSignal()
	rpcExitChan := make(chan error, 0)
	go handler.StartRpcServer(a, grpcAddress, rpcExitChan)

	select {
	case err := <-rpcExitChan:
		log.GetDefaultLogger().Fatalf("Grpc existing unexpected: %v", err)
	case <-serverExitChan:
		log.GetDefaultLogger().Info("Agent Existing")
	}
	return nil
}
