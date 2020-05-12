// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// GetV1SupportIPHandlerFunc turns a function with the right signature into a get v1 support IP handler
type GetV1SupportIPHandlerFunc func(GetV1SupportIPParams) middleware.Responder

// Handle executing the request and returning a response
func (fn GetV1SupportIPHandlerFunc) Handle(params GetV1SupportIPParams) middleware.Responder {
	return fn(params)
}

// GetV1SupportIPHandler interface for that can handle valid get v1 support IP params
type GetV1SupportIPHandler interface {
	Handle(GetV1SupportIPParams) middleware.Responder
}

// NewGetV1SupportIP creates a new http.Handler for the get v1 support IP operation
func NewGetV1SupportIP(ctx *middleware.Context, handler GetV1SupportIPHandler) *GetV1SupportIP {
	return &GetV1SupportIP{Context: ctx, Handler: handler}
}

/*GetV1SupportIP swagger:route GET /v1/support/ip getV1SupportIp

Returns this server's public ip.

*/
type GetV1SupportIP struct {
	Context *middleware.Context
	Handler GetV1SupportIPHandler
}

func (o *GetV1SupportIP) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetV1SupportIPParams()

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}
