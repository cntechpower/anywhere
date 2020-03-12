package main

import (
	"anywhere/log"
	"anywhere/server/anywhereServer"
	"anywhere/server/handler/rpcHandler"
	"anywhere/server/restapi/api/restapi"
	"anywhere/server/restapi/api/restapi/operations"
	"anywhere/tls"
	"anywhere/util"
	tls_ "crypto/tls"
	"fmt"
	"net"
	"net/http"

	"github.com/go-openapi/loads"

	"github.com/spf13/cobra"
)

//server global config
var version string

//args for add proxy config command
var addProxyAgentId, addProxyLocalAddr, addProxyWhiteListIps string
var addProxyRemoteAddr int
var addProxyIsWhiteListOn bool

//args for del proxy config command
var delProxyAgentId, delProxyLocalAddr string

func main() {
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
				panic(err)
			}
		},
	}
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
			if err := rpcHandler.ListAgent(); err != nil {
				fmt.Printf("error query agent list: %v\n", err)
			}
		},
	}
	var proxyCmd = &cobra.Command{
		Use:   "proxy",
		Short: "proxy admin interface",
		Long:  `proxy admin interface.`,
	}
	var proxyAddCmd = &cobra.Command{
		Use:   "add",
		Short: "add proxy config",
		Long:  `add a proxy config.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := rpcHandler.AddProxyConfig(addProxyAgentId, addProxyRemoteAddr, addProxyLocalAddr, addProxyIsWhiteListOn, addProxyWhiteListIps); err != nil {
				fmt.Printf("error adding proxy config : %v\n", err)
			}
		},
	}

	var proxyDelCmd = &cobra.Command{
		Use:   "del",
		Short: "delete proxy config",
		Long:  `delete a proxy config.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := rpcHandler.RemoveProxyConfig(delProxyAgentId, delProxyLocalAddr); err != nil {
				fmt.Printf("error deleting proxy config : %v\n", err)
			}
		},
	}

	var proxyListCmd = &cobra.Command{
		Use:   "list",
		Short: "list proxy configs",
		Long:  `add a proxy config.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := rpcHandler.ListProxyConfigs(); err != nil {
				fmt.Printf("error query proxy config list: %v\n", err)
			}
		},
	}

	var proxyLoadCmd = &cobra.Command{
		Use:   "load",
		Short: "load proxy configs",
		Long:  `load proxy configs from config file.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := rpcHandler.LoadProxyConfigFile(); err != nil {
				fmt.Printf("error load proxy config: %v\n", err)
			}
		},
	}

	var proxySaveCmd = &cobra.Command{
		Use:   "save",
		Short: "save proxy configs",
		Long:  `save proxy configs to config file.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := rpcHandler.SaveProxyConfigToFile(); err != nil {
				fmt.Printf("error save proxy config: %v\n", err)
			}
		},
	}
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
			if err := anywhereServer.WriteInitConfigFile(); err != nil {
				fmt.Printf("error reset proxy config: %v\n", err)
			}
		},
	}

	proxyAddCmd.PersistentFlags().StringVar(&addProxyAgentId, "agent-id", "", "belong to which agent")
	proxyAddCmd.PersistentFlags().IntVar(&addProxyRemoteAddr, "remote-addr", 0, "remote port")
	proxyAddCmd.PersistentFlags().StringVar(&addProxyLocalAddr, "local-addr", "127.0.0.1:80", "local addr")
	proxyAddCmd.PersistentFlags().StringVar(&addProxyWhiteListIps, "white-list", "", "local port")
	proxyAddCmd.PersistentFlags().BoolVarP(&addProxyIsWhiteListOn, "enable-wl", "", true, "enable white list or not")

	proxyDelCmd.PersistentFlags().StringVar(&delProxyAgentId, "agent-id", "", "del from which agent")
	proxyDelCmd.PersistentFlags().StringVar(&delProxyLocalAddr, "local-addr", "", "del from which localAddr")

	//main service
	rootCmd.AddCommand(startCmd)
	//agent cmds
	rootCmd.AddCommand(agentCmd)
	agentCmd.AddCommand(agentListCmd)

	//proxy cmds
	rootCmd.AddCommand(proxyCmd)
	proxyCmd.AddCommand(proxyListCmd)
	proxyCmd.AddCommand(proxyAddCmd)
	proxyCmd.AddCommand(proxyDelCmd)
	proxyCmd.AddCommand(proxyLoadCmd)
	proxyCmd.AddCommand(proxySaveCmd)

	//config file manage cmds
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(resetConfigCmd)
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}

}

func run(_ *cobra.Command, _ []string) error {
	log.InitLogger("")
	c, err := anywhereServer.ParseSystemConfigFile()
	if err != nil {
		return err
	}
	s := anywhereServer.InitServerInstance(c.ServerId, c.Net.MainPort)

	tlsConfig, err := tls.ParseTlsConfig(c.Ssl.CertFile, c.Ssl.KeyFile, c.Ssl.CaFile)
	if err != nil {
		return err
	}
	s.SetCredentials(tlsConfig)

	//start main service
	s.Start()

	// start rpc server
	rpcExitChan := make(chan error, 0)
	go rpcHandler.StartRpcServer(c.Net.GrpcPort, rpcExitChan)
	webExitChan := make(chan error, 0)
	if c.Net.IsWebEnable {
		go startUIAndAPIService(c.Net.WebAddr, c.User.AdminUser, c.User.AdminPass, webExitChan)

	}

	if err := anywhereServer.WriteInitConfigFile(); err != nil {
		return err
	}

	//wait for os kill signal. TODO: graceful shutdown
	go util.ListenTTINSignalLoop()
	serverExitChan := util.ListenKillSignal()
	select {
	case <-serverExitChan:
		log.GetDefaultLogger().Infof("Server Existing")
	case err := <-webExitChan:
		log.GetDefaultLogger().Fatalf("api server exit with error: %v", err)
	case err := <-rpcExitChan:
		log.GetDefaultLogger().Fatalf("rpc server exit with error: %v", err)
	case err := <-s.ExitChan:
		log.GetDefaultLogger().Fatalf("anywhere server exit with error: %v", err)
	}
	return nil
}

func startAPIServer(addr string, tlsConfig *tls_.Config, errChan chan error) {
	if err := util.CheckAddrValid(addr); err != nil {
		errChan <- err
	}

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		errChan <- err
	}

	api := operations.NewAnywhereServerAPI(swaggerSpec)
	server := restapi.NewServer(api)
	defer server.Shutdown()
	server.ConfigureAPI()
	var l net.Listener
	if tlsConfig != nil {
		l, err = tls_.Listen("tcp", addr, tlsConfig)
	} else {
		l, err = net.Listen("tcp", addr)
	}
	if err != nil {
		errChan <- err
	}
	handler := server.GetHandler()
	if err := http.Serve(l, handler); err != nil {
		errChan <- err
	}

}
