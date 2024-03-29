// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"net/http"

	"github.com/cntechpower/anywhere/server/api/http/handler"

	"github.com/cntechpower/anywhere/server/api/http/api/models"
	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	"github.com/cntechpower/anywhere/server/api/http/api/restapi/operations"

	"github.com/rs/cors"
)

//go:generate swagger generate server --target ../../api --name AnywhereServer --spec ../../definition/anywhere.yml --exclude-main

func configureFlags(api *operations.AnywhereServerAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.AnywhereServerAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	//api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	api.GetV1AgentListHandler = operations.GetV1AgentListHandlerFunc(func(params operations.GetV1AgentListParams) middleware.Responder {
		res, err := handler.ListAgentV1()
		if err != nil {
			return operations.NewGetV1AgentListDefault(500).WithPayload(models.GenericErrors(err.Error()))
		} else {
			return operations.NewGetV1AgentListOK().WithPayload(res)
		}

	})

	api.GetV1ProxyListHandler = operations.GetV1ProxyListHandlerFunc(func(params operations.GetV1ProxyListParams) middleware.Responder {
		res, err := handler.ListProxyV1()
		if err != nil {
			return operations.NewGetV1ProxyListDefault(500).WithPayload(models.GenericErrors(err.Error()))
		} else {
			return operations.NewGetV1ProxyListOK().WithPayload(res)
		}
	})

	api.GetV1ZoneListHandler = operations.GetV1ZoneListHandlerFunc(func(params operations.GetV1ZoneListParams) middleware.Responder {
		res := handler.ListZonesV1(params)
		return operations.NewGetV1ZoneListOK().WithPayload(res)
	})

	api.PostV1ProxyAddHandler = operations.PostV1ProxyAddHandlerFunc(func(params operations.PostV1ProxyAddParams) middleware.Responder {
		res, err := handler.AddProxyConfigV1(params)
		if err != nil {
			return operations.NewPostV1ProxyAddDefault(500).WithPayload(models.GenericErrors(err.Error()))
		} else {
			return operations.NewPostV1ProxyAddOK().WithPayload(res)
		}
	})

	api.GetV1SupportIPHandler = operations.GetV1SupportIPHandlerFunc(func(params operations.GetV1SupportIPParams) middleware.Responder {
		res, err := handler.GetV1SupportIP(params)
		if err != nil {
			return operations.NewGetV1SupportIPDefault(500).WithPayload(models.GenericErrors(err.Error()))
		}
		return operations.NewGetV1SupportIPOK().WithPayload(res)
	})

	api.PostV1ProxyUpdateHandler = operations.PostV1ProxyUpdateHandlerFunc(func(params operations.PostV1ProxyUpdateParams) middleware.Responder {
		res, err := handler.UpdateProxyConfigV1(params)
		if err != nil {
			return operations.NewPostV1ProxyAddDefault(500).WithPayload(models.GenericErrors(err.Error()))
		}
		return operations.NewPostV1ProxyAddOK().WithPayload(res)
	})

	api.PostV1ProxyDeleteHandler = operations.PostV1ProxyDeleteHandlerFunc(func(params operations.PostV1ProxyDeleteParams) middleware.Responder {
		res, err := handler.PostV1ProxyDeleteHandler(params)
		if err != nil {
			return operations.NewPostV1ProxyDeleteDefault(500).WithPayload(models.GenericErrors(err.Error()))
		}
		return operations.NewPostV1ProxyDeleteOK().WithPayload(res)
	})

	api.GetV1SummaryHandler = operations.GetV1SummaryHandlerFunc(func(params operations.GetV1SummaryParams) middleware.Responder {
		res, err := handler.GetSummaryV1()
		if err != nil {
			return operations.NewGetV1SummaryDefault(500).WithPayload(models.GenericErrors(err.Error()))
		}
		return operations.NewGetV1SummaryOK().WithPayload(res)
	})

	api.GetV1ConnectionListHandler = operations.GetV1ConnectionListHandlerFunc(func(params operations.GetV1ConnectionListParams) middleware.Responder {
		res, err := handler.GetConnsV1(params)
		if err != nil {
			return operations.NewGetV1ConnectionListDefault(500).WithPayload(models.GenericErrors(err.Error()))
		}
		return operations.NewGetV1ConnectionListOK().WithPayload(res)
	})

	api.PostV1ConnectionKillHandler = operations.PostV1ConnectionKillHandlerFunc(func(params operations.PostV1ConnectionKillParams) middleware.Responder {
		res, err := handler.KillConnV1(params)
		if err != nil {
			return operations.NewPostV1ConnectionKillDefault(500).WithPayload(models.GenericErrors(err.Error()))
		}
		return operations.NewPostV1ConnectionKillOK().WithPayload(res)
	})

	api.GetV1WhitelistDenysHandler = operations.GetV1WhitelistDenysHandlerFunc(func(params operations.GetV1WhitelistDenysParams) middleware.Responder {
		res, err := handler.WhiteListRecordV1(params)
		if err != nil {
			return operations.NewGetV1WhitelistDenysDefault(500).WithPayload(models.GenericErrors(err.Error()))
		}
		return operations.NewGetV1WhitelistDenysOK().WithPayload(res)
	})

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix"
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return cors.AllowAll().Handler(handler)
}
