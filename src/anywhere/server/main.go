package main

import (
	"anywhere/log"
	"anywhere/server/anywhereServer"
	"anywhere/server/restapi/apiServer"
	"anywhere/tls"
	"anywhere/util"

	"github.com/spf13/cobra"
)

var port, grpcPort int
var certFile, keyFile, caFile, serverId string

func main() {
	var rootCmd = &cobra.Command{
		Use:   "anywhered --help",
		Short: "This is A Proxy Server ",
		Long:  `anywhere Version 0.0.1`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := run(cmd, args); err != nil {
				panic(err)
			}
		},
	}
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 1111, "anywhered serve port")
	rootCmd.PersistentFlags().IntVarP(&grpcPort, "grpc-port", "g", 1112, "anywhered grpc port")
	rootCmd.PersistentFlags().StringVarP(&serverId, "server-id", "s", "anywhered-1", "anywhered server id")
	rootCmd.PersistentFlags().StringVar(&certFile, "cert", "../../credential/server.crt", "cert file")
	rootCmd.PersistentFlags().StringVar(&keyFile, "key", "../../credential/server.key", "key file")
	rootCmd.PersistentFlags().StringVar(&caFile, "ca", "../../credential/ca.crt", "ca file")
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
	rpcExitChan := make(chan error, 0)
	go apiServer.StartAPIServer(grpcPort, tlsConfig, rpcExitChan)

	select {
	case <-serverExitChan:
		log.Info("Server Existing")
		s.ListAgentInfo()
		s.ListProxyConfig()
	case err := <-rpcExitChan:
		log.Fatal("rpc exit with error: %v", err)
	case err := <-s.ExitChan:
		panic(err)

	}
	return nil
}
