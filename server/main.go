package main

import (
	"time"

	"github.com/cntechpower/anywhere/server/db"

	"github.com/cntechpower/anywhere/server/restapi/api/restapi"
	"github.com/cntechpower/anywhere/server/restapi/api/restapi/operations"
	"github.com/go-openapi/loads"

	"github.com/cntechpower/anywhere/server/http"

	"github.com/cntechpower/anywhere/server/conf"

	"github.com/cntechpower/anywhere/server/cmd"
	"github.com/cntechpower/anywhere/server/rpc/handler"
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

	//proxy cmd
	rootCmd.AddCommand(cmd.Proxy())

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
	conf.Init()
	s := server.InitServerInstance(conf.Conf.ServerId, conf.Conf.MainPort, conf.Conf.User)
	tlsConfig, err := tls.ParseTlsConfig(conf.Conf.AgentSsl.CertFile, conf.Conf.AgentSsl.KeyFile, conf.Conf.AgentSsl.CaFile)
	if err != nil {
		return err
	}
	s.SetCredentials(tlsConfig)

	//start main service
	s.Start()

	// start rpc server
	rpcExitChan := make(chan error, 0)
	go handler.StartRpcServer(s, conf.Conf.UiConfig.GrpcAddr, rpcExitChan)
	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		return err
	}
	api := operations.NewAnywhereServerAPI(swaggerSpec)
	restServer := restapi.NewServer(api)
	restServer.ConfigureAPI()
	restHandler := restServer.GetHandler()
	webExitChan := make(chan error, 0)
	//TODO: passthroughs config directly to http.StartUIAndAPIService
	if conf.Conf.UiConfig.IsWebEnable {
		go http.StartUIAndAPIService(restHandler, s, conf.Conf.UiConfig.WebAddr, webExitChan,
			conf.Conf.UiConfig.SkipLogin, conf.Conf.UiConfig.DebugMode, conf.Conf.ReportWhiteCidrs)

	}

	//wait for os kill signal. TODO: graceful shutdown
	go os.ListenTTINSignalLoop()
	//delay init of persist
	go func() {
		time.Sleep(5 * time.Second)
		db.Init(conf.Conf.MysqlDSN)
	}()
	defer db.Close()
	serverExitChan := os.ListenKillSignal()
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
