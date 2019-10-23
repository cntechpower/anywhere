package apiServer

import (
	"anywhere/server/restapi/api/restapi"
	"anywhere/server/restapi/api/restapi/operations"

	"github.com/go-openapi/loads"
)

func StartAPIServer(port int, errChan chan error) {
	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		errChan <- err
	}

	api := operations.NewAnywhereServerAPI(swaggerSpec)
	server := restapi.NewServer(api)
	defer server.Shutdown()
	server.Port = port
	server.ConfigureAPI()

	if err := server.Serve(); err != nil {
		errChan <- err
	}

}
