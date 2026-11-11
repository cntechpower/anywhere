package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/cntechpower/utils/log"
	xos "github.com/cntechpower/utils/os"
	"github.com/cntechpower/utils/tracing"

	"github.com/cntechpower/anywhere/dao"
	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/server/api"
	"github.com/cntechpower/anywhere/server/cmd"
	"github.com/cntechpower/anywhere/server/conf"
	"github.com/cntechpower/anywhere/server/server"
	tlsutil "github.com/cntechpower/anywhere/tls"

	"github.com/spf13/cobra"
)

const (
	app = "main.anywhered"
	// shutdownTimeout bounds the graceful shutdown of listeners and tunnels.
	shutdownTimeout = 10 * time.Second
)

// server global config
var version string

func main() {
	// init log
	esAddr := os.Getenv("ES_ADDR")
	logOptions := make([]log.Option, 0)
	logOptions = append(logOptions, log.WithStd(log.OutputTypeText))
	if esAddr != "" {
		logOptions = append(logOptions, log.WithEs(app, esAddr))
	}
	log.Init(logOptions...)
	defer log.Close()

	dao.Init(model.GetPersistModels())
	defer dao.Close()

	// init trace
	traceAddr := os.Getenv("TRACE_ADDR")
	if traceAddr != "" {
		tracing.Init(app, traceAddr)
		defer tracing.Close()
	}

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
				log.Fatalf(log.NewHeader("serverMain"), "%s", err.Error())
			}
		},
	}

	// main service
	rootCmd.AddCommand(startCmd)

	// agent cmd
	rootCmd.AddCommand(cmd.Agent())

	// config file manage cmd
	rootCmd.AddCommand(cmd.Config())

	// conn cmd
	rootCmd.AddCommand(cmd.Conn())

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}

}

func run(_ *cobra.Command, _ []string) error {
	h := log.NewHeader("serverMain")
	if err := conf.Init(); err != nil {
		return err
	}
	s, err := server.New(
		conf.Conf.ServerId,
		conf.Conf.MainPort,
		conf.Conf.User,
		time.Duration(conf.Conf.ProxyConnIdleTimeoutSeconds)*time.Second,
	)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tlsConfig, err := tlsutil.ParseServerTlsConfig(conf.Conf.AgentSsl.CertFile, conf.Conf.AgentSsl.KeyFile, conf.Conf.AgentSsl.CaFile)
	if err != nil {
		return err
	}
	s.SetCredentials(tlsConfig)

	if err := s.Start(ctx); err != nil {
		return err
	}

	// start api
	apiExitChan := make(chan error, 2)
	if err := api.Start(ctx, s, tlsConfig, apiExitChan); err != nil {
		return err
	}

	go xos.ListenTTINSignalLoop()
	serverExitChan := xos.ListenKillSignal()

	select {
	case <-serverExitChan:
		h.Infof("shutdown signal received")
	case err := <-apiExitChan:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			h.Errorf("api server exited unexpectedly: %v", err)
		}
	}

	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()
	if err := s.Shutdown(shutdownCtx); err != nil {
		h.Errorf("graceful shutdown error: %v", err)
		return err
	}
	h.Infof("server stopped")
	return nil
}
