package api

import (
	"context"
	"crypto/tls"

	"github.com/cntechpower/anywhere/server/api/auth"
	httpHandler "github.com/cntechpower/anywhere/server/api/http/handler"
	rpcHandler "github.com/cntechpower/anywhere/server/api/rpc/handler"
	"github.com/cntechpower/anywhere/server/conf"
	"github.com/cntechpower/anywhere/server/server"
)

// Start wires the management gRPC server and the HTTP console. Every
// dependency is passed explicitly; there is no package level server instance.
func Start(ctx context.Context, s *server.Server, tlsConfig *tls.Config, apiExitChan chan error) error {
	go rpcHandler.StartRpcServer(ctx, s, conf.Conf.UiConfig.GrpcAddr, tlsConfig, apiExitChan)

	jwtValidator, err := auth.NewJwtValidator(conf.JwtSecret())
	if err != nil {
		return err
	}
	h := httpHandler.New(s, s.GetUserValidator(), jwtValidator)
	restHandler, err := httpHandler.NewRESTHandler(h)
	if err != nil {
		return err
	}
	go h.StartUIAndAPIService(ctx, restHandler, apiExitChan)
	return nil
}
