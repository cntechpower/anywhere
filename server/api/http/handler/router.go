package handler

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	"github.com/cntechpower/anywhere/server/api/http/api/models"
	"github.com/cntechpower/anywhere/server/api/http/api/restapi"
	"github.com/cntechpower/anywhere/server/api/http/api/restapi/operations"
)

// NewRESTHandler wires the generated Swagger operations to an API instance.
//
// The generated restapi/configure_anywhere_server.go file is intentionally not
// used: it would need package level handlers. Wiring here keeps everything
// dependency injected and testable.
func NewRESTHandler(h *API) (http.Handler, error) {
	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		return nil, err
	}
	api := operations.NewAnywhereServerAPI(swaggerSpec)
	api.ServeError = errors.ServeError
	api.JSONProducer = runtime.JSONProducer()

	api.GetV1AgentListHandler = operations.GetV1AgentListHandlerFunc(func(params operations.GetV1AgentListParams) middleware.Responder {
		res, err := h.ListAgentV1()
		if err != nil {
			return operations.NewGetV1AgentListDefault(500).WithPayload(models.GenericErrors(err.Error()))
		}
		return operations.NewGetV1AgentListOK().WithPayload(res)
	})

	api.GetV1ProxyListHandler = operations.GetV1ProxyListHandlerFunc(func(params operations.GetV1ProxyListParams) middleware.Responder {
		res, err := h.ListProxyV1()
		if err != nil {
			return operations.NewGetV1ProxyListDefault(500).WithPayload(models.GenericErrors(err.Error()))
		}
		return operations.NewGetV1ProxyListOK().WithPayload(res)
	})

	api.GetV1ZoneListHandler = operations.GetV1ZoneListHandlerFunc(func(params operations.GetV1ZoneListParams) middleware.Responder {
		return operations.NewGetV1ZoneListOK().WithPayload(h.ListZonesV1(params))
	})

	api.PostV1ProxyAddHandler = operations.PostV1ProxyAddHandlerFunc(func(params operations.PostV1ProxyAddParams) middleware.Responder {
		res, err := h.AddProxyConfigV1(params)
		if err != nil {
			return operations.NewPostV1ProxyAddDefault(500).WithPayload(models.GenericErrors(err.Error()))
		}
		return operations.NewPostV1ProxyAddOK().WithPayload(res)
	})

	api.GetV1SupportIPHandler = operations.GetV1SupportIPHandlerFunc(func(params operations.GetV1SupportIPParams) middleware.Responder {
		res, err := h.GetV1SupportIP(params)
		if err != nil {
			return operations.NewGetV1SupportIPDefault(500).WithPayload(models.GenericErrors(err.Error()))
		}
		return operations.NewGetV1SupportIPOK().WithPayload(res)
	})

	api.PostV1ProxyUpdateHandler = operations.PostV1ProxyUpdateHandlerFunc(func(params operations.PostV1ProxyUpdateParams) middleware.Responder {
		res, err := h.UpdateProxyConfigV1(params)
		if err != nil {
			return operations.NewPostV1ProxyUpdateDefault(500).WithPayload(models.GenericErrors(err.Error()))
		}
		return operations.NewPostV1ProxyUpdateOK().WithPayload(res)
	})

	api.PostV1ProxyDeleteHandler = operations.PostV1ProxyDeleteHandlerFunc(func(params operations.PostV1ProxyDeleteParams) middleware.Responder {
		res, err := h.PostV1ProxyDeleteHandler(params)
		if err != nil {
			return operations.NewPostV1ProxyDeleteDefault(500).WithPayload(models.GenericErrors(err.Error()))
		}
		return operations.NewPostV1ProxyDeleteOK().WithPayload(res)
	})

	api.GetV1SummaryHandler = operations.GetV1SummaryHandlerFunc(func(params operations.GetV1SummaryParams) middleware.Responder {
		res, err := h.GetSummaryV1()
		if err != nil {
			return operations.NewGetV1SummaryDefault(500).WithPayload(models.GenericErrors(err.Error()))
		}
		return operations.NewGetV1SummaryOK().WithPayload(res)
	})

	api.GetV1ConnectionListHandler = operations.GetV1ConnectionListHandlerFunc(func(params operations.GetV1ConnectionListParams) middleware.Responder {
		res, err := h.GetConnsV1(params)
		if err != nil {
			return operations.NewGetV1ConnectionListDefault(500).WithPayload(models.GenericErrors(err.Error()))
		}
		return operations.NewGetV1ConnectionListOK().WithPayload(res)
	})

	api.PostV1ConnectionKillHandler = operations.PostV1ConnectionKillHandlerFunc(func(params operations.PostV1ConnectionKillParams) middleware.Responder {
		res, err := h.KillConnV1(params)
		if err != nil {
			return operations.NewPostV1ConnectionKillDefault(500).WithPayload(models.GenericErrors(err.Error()))
		}
		return operations.NewPostV1ConnectionKillOK().WithPayload(res)
	})

	api.GetV1WhitelistDenysHandler = operations.GetV1WhitelistDenysHandlerFunc(func(params operations.GetV1WhitelistDenysParams) middleware.Responder {
		res, err := h.WhiteListRecordV1(params)
		if err != nil {
			return operations.NewGetV1WhitelistDenysDefault(500).WithPayload(models.GenericErrors(err.Error()))
		}
		return operations.NewGetV1WhitelistDenysOK().WithPayload(res)
	})

	api.ServerShutdown = func() {}

	return api.Serve(nil), nil
}
