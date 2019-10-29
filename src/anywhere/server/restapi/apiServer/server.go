package apiServer

import (
	"anywhere/server/restapi/api/restapi"
	"anywhere/server/restapi/api/restapi/operations"
	"anywhere/util"
	"crypto/tls"
	"net/http"

	"github.com/go-openapi/loads"
)

func StartAPIServer(port int, tlsConfig *tls.Config, errChan chan error) {
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
	l, err := tls.Listen("tcp", addr.String(), tlsConfig)
	if err != nil {
		errChan <- err
	}
	handler := server.GetHandler()
	if err := http.Serve(l, handler); err != nil {
		errChan <- err
	}

}
