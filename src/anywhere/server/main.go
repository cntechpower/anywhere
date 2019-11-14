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
	"net/http"

	"github.com/go-openapi/loads"

	"github.com/spf13/cobra"
)

var port, apiPort, grpcPort int
var certFile, keyFile, caFile, serverId string

var addProxyAgentId, addProxyRemotePort, addProxyLocalIp, addProxyLocalPort string

var delProxyAgentId, delProxyLocalIp, delProxyLocalPort string

func main() {
	var rootCmd = &cobra.Command{
		Use:   "anywhered",
		Short: "This is A Proxy Server ",
		Long:  `anywhere server Version 0.0.1`,
	}
	var startCmd = &cobra.Command{
		Use:   "start",
		Short: "start anywhered service",
		Long:  `anywhere server Version 0.0.1`,
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
			if err := rpcHandler.AddProxyConfig(grpcPort, addProxyAgentId, addProxyRemotePort, addProxyLocalIp, addProxyLocalPort); err != nil {
				fmt.Printf("error adding proxy config : %v\n", err)
			}
		},
	}

	var proxyDelCmd = &cobra.Command{
		Use:   "del",
		Short: "delete proxy config",
		Long:  `delete a proxy config.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := rpcHandler.RemoveProxyConfig(grpcPort, delProxyAgentId, delProxyLocalIp, delProxyLocalPort); err != nil {
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

	proxyAddCmd.PersistentFlags().StringVar(&addProxyAgentId, "agent-id", "", "belong to which agent")
	proxyAddCmd.PersistentFlags().StringVar(&addProxyRemotePort, "remote-port", "", "remote port")
	proxyAddCmd.PersistentFlags().StringVar(&addProxyLocalIp, "local-ip", "127.0.0.1", "local ip")
	proxyAddCmd.PersistentFlags().StringVar(&addProxyLocalPort, "local-port", "", "local port")
	proxyDelCmd.PersistentFlags().StringVar(&delProxyAgentId, "agent-id", "", "del from which agent")
	proxyDelCmd.PersistentFlags().StringVar(&delProxyLocalIp, "local-ip", "", "del from which localIp")
	proxyDelCmd.PersistentFlags().StringVar(&delProxyLocalPort, "local-port", "", "del from which localPort")
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 1111, "anywhered serve port")
	rootCmd.PersistentFlags().IntVarP(&apiPort, "api-port", "a", 1112, "anywhered rest api port")
	rootCmd.PersistentFlags().IntVarP(&grpcPort, "grpc-port", "g", 1113, "anywhered grpc port")
	rootCmd.PersistentFlags().StringVarP(&serverId, "server-id", "s", "anywhered-1", "anywhered server id")
	rootCmd.PersistentFlags().StringVar(&certFile, "cert", "credential/server.crt", "cert file")
	rootCmd.PersistentFlags().StringVar(&keyFile, "key", "credential/server.key", "key file")
	rootCmd.PersistentFlags().StringVar(&caFile, "ca", "credential/ca.crt", "ca file")

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

	// start api server
	apiExitChan := make(chan error, 0)
	//go apiServer.StartAPIServer(apiPort, tlsConfig, apiExitChan)
	//restHandler.StartAPIServer(apiPort, tlsConfig, apiExitChan)
	go startAPIServer(apiPort, tlsConfig, apiExitChan)

	// start rpc server
	rpcExitChan := make(chan error, 0)
	//go rpcServer.StartRpcServer(grpcPort, rpcExitChan)
	rpcHandler.StartRpcServer(grpcPort, rpcExitChan)

	//wait for os kill signal. TODO: graceful shutdown
	go util.ListenTTINSignal()
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

func startAPIServer(port int, tlsConfig *tls_.Config, errChan chan error) {
	addr, err := util.GetAddrByIpPort("127.0.0.1", port)
	if err != nil {
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
	l, err := tls_.Listen("tcp", addr.String(), tlsConfig)
	if err != nil {
		errChan <- err
	}
	handler := server.GetHandler()
	if err := http.Serve(l, handler); err != nil {
		errChan <- err
	}

}
