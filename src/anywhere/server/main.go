package main

import (
	"anywhere/log"
	"anywhere/server/anywhereServer"
	"anywhere/server/restapi/apiServer"
	"anywhere/server/rpc/rpcServer"
	"anywhere/tls"
	"anywhere/util"
	"fmt"

	"github.com/spf13/cobra"
)

var port, apiPort, grpcPort int
var certFile, keyFile, caFile, serverId string

var addProxyAgentId, addProxyRemotePort, addProxyLocalIp, addProxyLocalPort string

func main() {
	var rootCmd = &cobra.Command{
		Use:   "anywhered",
		Short: "This is A Proxy Server ",
		Long:  `anywhere server Version 0.0.1`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := run(cmd, args); err != nil {
				panic(err)
			}
		},
	}
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "list agents",
		Long:  `list anywhere agetns.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := rpcServer.RpcListAgent(grpcPort); err != nil {
				fmt.Printf("error query agent list: %v\n", err)
			}
		},
	}
	var addProxyCmd = &cobra.Command{
		Use:   "add",
		Short: "add",
		Long:  `add a proxy config.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := rpcServer.RpcAddProxyConfig(grpcPort, addProxyAgentId, addProxyRemotePort, addProxyLocalIp, addProxyLocalPort); err != nil {
				fmt.Printf("error query agent list: %v\n", err)
			}
		},
	}
	addProxyCmd.PersistentFlags().StringVar(&addProxyAgentId, "agent-id", "", "belong to which agent")
	addProxyCmd.PersistentFlags().StringVar(&addProxyRemotePort, "remote-port", "", "remote port")
	addProxyCmd.PersistentFlags().StringVar(&addProxyLocalIp, "local-ip", "127.0.0.1", "local ip")
	addProxyCmd.PersistentFlags().StringVar(&addProxyLocalPort, "local-port", "", "local port")
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 1111, "anywhered serve port")
	rootCmd.PersistentFlags().IntVarP(&apiPort, "api-port", "a", 1112, "anywhered rest api port")
	rootCmd.PersistentFlags().IntVarP(&grpcPort, "grpc-port", "g", 1113, "anywhered grpc port")
	rootCmd.PersistentFlags().StringVarP(&serverId, "server-id", "s", "anywhered-1", "anywhered server id")
	rootCmd.PersistentFlags().StringVar(&certFile, "cert", "credential/server.crt", "cert file")
	rootCmd.PersistentFlags().StringVar(&keyFile, "key", "credential/server.key", "key file")
	rootCmd.PersistentFlags().StringVar(&caFile, "ca", "credential/ca.crt", "ca file")
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(addProxyCmd)
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}

}

func run(_ *cobra.Command, _ []string) error {
	log.InitStdLogger()
	s := anywhereServer.InitServerInstance(serverId, port)

	tlsConfig, err := tls.ParseTlsConfig(certFile, keyFile, caFile)
	if err != nil {
		return err
	}
	s.SetCredentials(tlsConfig)

	s.Start()
	serverExitChan := util.ListenKillSignal()
	apiExitChan := make(chan error, 0)
	go apiServer.StartAPIServer(apiPort, tlsConfig, apiExitChan)
	rpcExitChan := make(chan error, 0)
	go rpcServer.StartRpcServer(grpcPort, rpcExitChan)
	select {
	case <-serverExitChan:
		log.Info("Server Existing")
		s.ListAgentInfo()
		s.ListProxyConfig()
	case err := <-apiExitChan:
		log.Fatal("api server exit with error: %v", err)
	case err := <-rpcExitChan:
		log.Fatal("rpc server exit with error: %v", err)
	case err := <-s.ExitChan:
		log.Fatal("anywhere server exit with error: %v", err)

	}
	return nil
}
