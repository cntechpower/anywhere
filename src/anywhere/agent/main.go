package main

import (
	"anywhere/agent/anywhereAgent"
	"anywhere/log"
	"anywhere/util"

	"github.com/spf13/cobra"
)

var serverPort int
var serverIp, agentId, certFile, keyFile, caFile string

func main() {
	var rootCmd = &cobra.Command{
		Use:   "anywhere --help",
		Short: "This is A Proxy Agent ",
		Long:  `anywhere agent Version 0.0.1`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := run(cmd, args); err != nil {
				panic(err)
			}
		},
	}
	rootCmd.PersistentFlags().StringVarP(&serverIp, "server-ip", "s", "127.0.0.1", "anywhered server address")
	rootCmd.PersistentFlags().IntVarP(&serverPort, "server-port", "p", 1111, "anywhered server port")
	rootCmd.PersistentFlags().StringVarP(&agentId, "agent-id", "i", "anywhere-agent-1", "anywhere agent id")
	rootCmd.PersistentFlags().StringVar(&certFile, "cert", "credential/client.crt", "cert file")
	rootCmd.PersistentFlags().StringVar(&keyFile, "key", "credential/client.key", "key file")
	rootCmd.PersistentFlags().StringVar(&caFile, "ca", "credential/ca.crt", "ca file")
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

	go util.ListenTTINSignal()
	serverExitChan := util.ListenKillSignal()

	select {
	case <-serverExitChan:
		log.GetDefaultLogger().Info("Agent Existing")
	}
	return nil
}
