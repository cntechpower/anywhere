package api

import (
	"github.com/cntechpower/anywhere/server/api/http/api/restapi"
	"github.com/cntechpower/anywhere/server/api/http/api/restapi/operations"
	httpHandler "github.com/cntechpower/anywhere/server/api/http/handler"
	rpcHandler "github.com/cntechpower/anywhere/server/api/rpc/handler"
	"github.com/cntechpower/anywhere/server/conf"
	"github.com/cntechpower/anywhere/server/server"
	"github.com/go-openapi/loads"
)

func Start(s *server.Server, apiExitChan chan error) (err error) {
	go rpcHandler.StartRpcServer(s, conf.Conf.UiConfig.GrpcAddr, apiExitChan)

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		return err
	}
	api := operations.NewAnywhereServerAPI(swaggerSpec)
	restServer := restapi.NewServer(api)
	restServer.ConfigureAPI()
	restHandler := restServer.GetHandler()
	go httpHandler.StartUIAndAPIService(restHandler, s, apiExitChan)
	return
}
