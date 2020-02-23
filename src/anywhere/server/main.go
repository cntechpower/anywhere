package main

import (
	"anywhere/frontEnd"
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
var port, grpcPort int
var certFile, keyFile, caFile, serverId string

//web interface config
var isWebEnable bool
var webAddress, restAddress string

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
			if err := rpcHandler.ListAgent(grpcPort); err != nil {
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
			if err := rpcHandler.AddProxyConfig(grpcPort, addProxyAgentId, addProxyRemoteAddr, addProxyLocalAddr, addProxyIsWhiteListOn, addProxyWhiteListIps); err != nil {
				fmt.Printf("error adding proxy config : %v\n", err)
			}
		},
	}

	var proxyDelCmd = &cobra.Command{
		Use:   "del",
		Short: "delete proxy config",
		Long:  `delete a proxy config.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := rpcHandler.RemoveProxyConfig(grpcPort, delProxyAgentId, delProxyLocalAddr); err != nil {
				fmt.Printf("error deleting proxy config : %v\n", err)
			}
		},
	}

	var proxyListCmd = &cobra.Command{
		Use:   "list",
		Short: "list proxy configs",
		Long:  `add a proxy config.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := rpcHandler.ListProxyConfigs(grpcPort); err != nil {
				fmt.Printf("error query proxy config list: %v\n", err)
			}
		},
	}

	var proxyLoadCmd = &cobra.Command{
		Use:   "load",
		Short: "load proxy configs",
		Long:  `load proxy configs from config file.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := rpcHandler.LoadProxyConfigFile(grpcPort); err != nil {
				fmt.Printf("error load proxy config: %v\n", err)
			}
		},
	}

	var proxySaveCmd = &cobra.Command{
		Use:   "save",
		Short: "save proxy configs",
		Long:  `save proxy configs to config file.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := rpcHandler.SaveProxyConfigToFile(grpcPort); err != nil {
				fmt.Printf("error save proxy config: %v\n", err)
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

	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 1111, "anywhered serve port")
	rootCmd.PersistentFlags().StringVarP(&restAddress, "rest-address", "a", "127.0.0.1:1112", "anywhered rest api address")
	rootCmd.PersistentFlags().IntVarP(&grpcPort, "grpc-port", "g", 1113, "anywhered grpc port")
	rootCmd.PersistentFlags().StringVarP(&serverId, "server-id", "s", "anywhered-1", "anywhered server id")
	rootCmd.PersistentFlags().StringVar(&certFile, "cert", "credential/server.crt", "cert file")
	rootCmd.PersistentFlags().StringVar(&keyFile, "key", "credential/server.key", "key file")
	rootCmd.PersistentFlags().StringVar(&caFile, "ca", "credential/ca.crt", "ca file")
	rootCmd.PersistentFlags().BoolVar(&isWebEnable, "web", false, "enable web interface")
	rootCmd.PersistentFlags().StringVar(&webAddress, "web-address", "0.0.0.0:1114", "web interface port")

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
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}

}

func run(_ *cobra.Command, _ []string) error {
	log.InitLogger("")
	s := anywhereServer.InitServerInstance(serverId, port)

	tlsConfig, err := tls.ParseTlsConfig(certFile, keyFile, caFile)
	if err != nil {
		return err
	}
	s.SetCredentials(tlsConfig)

	//start main service
	s.Start()

	// start rpc server
	rpcExitChan := make(chan error, 0)
	//go rpcServer.StartRpcServer(grpcPort, rpcExitChan)
	go rpcHandler.StartRpcServer(grpcPort, rpcExitChan)
	apiExitChan := make(chan error, 0)
	webExitChan := make(chan struct{}, 0)
	if isWebEnable {
		// start api server
		//go startAPIServer(apiPort, tlsConfig, apiExitChan)
		go startAPIServer(restAddress, nil, apiExitChan) //TODO: swagger with tls
		//start web interface

		go frontEnd.Start(webAddress, webExitChan)
	}

	//wait for os kill signal. TODO: graceful shutdown
	go util.ListenTTINSignalLoop()
	serverExitChan := util.ListenKillSignal()
	select {
	case <-serverExitChan:
		log.GetDefaultLogger().Infof("Server Existing")
	case err := <-apiExitChan:
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
	server.Port = port
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
