package main

import (
	"context"

	"github.com/cntechpower/anywhere/server/api"

	"github.com/cntechpower/anywhere/dao"
	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/server/cmd"
	"github.com/cntechpower/anywhere/server/conf"
	"github.com/cntechpower/anywhere/server/server"
	"github.com/cntechpower/anywhere/tls"
	"github.com/cntechpower/utils/log"
	"github.com/cntechpower/utils/os"

	"github.com/spf13/cobra"
)

const (
	app = "main.anywhered"
)

//server global config
var version string

func main() {
	log.Init(
		log.WithStd(log.OutputTypeText),
		log.WithEs(app, "http://127.0.0.1:9200"),
	)
	defer log.Close()
	conf.Init()
	dao.Init(conf.Conf.MysqlDSN, model.GetPersistModels(), model.GetTmpModels())
	defer dao.Close()

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

	//agent cmd
	rootCmd.AddCommand(cmd.Agent())

	//config file manage cmd
	rootCmd.AddCommand(cmd.Config())

	//conn cmd
	rootCmd.AddCommand(cmd.Conn())

	//status cmd
	rootCmd.AddCommand(cmd.Status())
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}

}

func run(_ *cobra.Command, _ []string) error {
	h := log.NewHeader("serverMain")
	s := server.InitServerInstance(conf.Conf.ServerId, conf.Conf.MainPort, conf.Conf.User)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tlsConfig, err := tls.ParseTlsConfig(conf.Conf.AgentSsl.CertFile, conf.Conf.AgentSsl.KeyFile, conf.Conf.AgentSsl.CaFile)
	if err != nil {
		return err
	}
	s.SetCredentials(tlsConfig)

	//start main service
	s.Start(ctx)

	// start api
	apiExitChan := make(chan error, 0)
	err = api.Start(s, apiExitChan)
	if err != nil {
		panic(err)
	}

	//wait for os kill signal. TODO: graceful shutdown
	go os.ListenTTINSignalLoop()

	serverExitChan := os.ListenKillSignal()
	select {
	case <-serverExitChan:
		log.Infof(h, "Server Existing")
	case err := <-apiExitChan:
		log.Fatalf(h, "api server exit with error: %v", err)
	case err := <-s.ExitChan:
		log.Fatalf(h, "anywhere server exit with error: %v", err)
	}
	return nil
}
