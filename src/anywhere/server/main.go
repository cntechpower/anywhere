package main

import (
	"anywhere/log"
	"anywhere/server/anywhereServer"
	"anywhere/server/cmd"
	"anywhere/server/handler/rpcHandler"
	"anywhere/tls"
	"anywhere/util"

	"github.com/spf13/cobra"
)

//server global config
var version string

func main() {
	log.InitLogger("")
	var rootCmd = &cobra.Command{
		Use:   "anywhered",
		Short: "This is A Proxy Server ",
		Long:  "anywhere server - " + version,
	}
	var startCmd = &cobra.Command{
		Use:   "start",
		Short: "start anywhered service",
		Long:  "anywhere server Version 0.0.1 -" + version,
		Run: func(cmd *cobra.Command, args []string) {
			if err := run(cmd, args); err != nil {
				log.Fatalf(log.NewHeader("serverMain"), err.Error())
			}
		},
	}

	//main service
	rootCmd.AddCommand(startCmd)
	//agent cmds
	rootCmd.AddCommand(cmd.GetAgentCmd())

	//proxy cmds
	rootCmd.AddCommand(cmd.GetProxyCmd())

	//config file manage cmds
	rootCmd.AddCommand(cmd.GetConfigCmd())

	//conn cmds
	rootCmd.AddCommand(cmd.GetConnCmd())
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}

}

func run(_ *cobra.Command, _ []string) error {
	h := log.NewHeader("serverMain")
	c, err := anywhereServer.ParseSystemConfigFile()
	if err != nil {
		return err
	}
	s := anywhereServer.InitServerInstance(c.ServerId, c.MainPort)
	tlsConfig, err := tls.ParseTlsConfig(c.Ssl.CertFile, c.Ssl.KeyFile, c.Ssl.CaFile)
	if err != nil {
		return err
	}
	s.SetCredentials(tlsConfig)

	//start main service
	s.Start()

	// start rpc server
	rpcExitChan := make(chan error, 0)
	go rpcHandler.StartRpcServer(s, c.UiConfig.GrpcAddr, rpcExitChan)
	webExitChan := make(chan error, 0)
	if c.UiConfig.IsWebEnable {
		go startUIAndAPIService(c.UiConfig.WebAddr, c.User.AdminUser, c.User.AdminPass, c.User.AdminOtpCode, c.User.AdminOtpEnable, webExitChan, c.UiConfig.SkipLogin, c.UiConfig.DebugMode)

	}

	//wait for os kill signal. TODO: graceful shutdown
	go util.ListenTTINSignalLoop()
	serverExitChan := util.ListenKillSignal()
	select {
	case <-serverExitChan:
		log.Infof(h, "Server Existing")
	case err := <-webExitChan:
		log.Fatalf(h, "api server exit with error: %v", err)
	case err := <-rpcExitChan:
		log.Fatalf(h, "rpc server exit with error: %v", err)
	case err := <-s.ExitChan:
		log.Fatalf(h, "anywhere server exit with error: %v", err)
	}
	return nil
}
