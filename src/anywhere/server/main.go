package main

import (
	"anywhere/log"
	"anywhere/server/anywhereServer"
	"anywhere/util"

	"github.com/spf13/cobra"
)

var port int
var certFile, keyFile, caFile string

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
	rootCmd.PersistentFlags().StringVar(&certFile, "cert", "../credential/server.crt", "cert file")
	rootCmd.PersistentFlags().StringVar(&keyFile, "key", "../credential/server.key", "key file")
	rootCmd.PersistentFlags().StringVar(&caFile, "ca", "../credential/ca.crt", "ca file")
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}

}

func run(_ *cobra.Command, _ []string) error {
	log.InitStdLogger()
	s := anywhereServer.InitServerInstance("server-id", port, true, true)
	if err := s.SetCredentials(certFile, keyFile, caFile); err != nil {
		return err
	}
	s.Start()
	serverExitChan := util.ListenKillSignal()

	select {
	case <-serverExitChan:
		log.Info("Server Existing")
		s.ListAgentInfo()
	case err := <-s.ExitChan:
		panic(err)

	}
	return nil
}
