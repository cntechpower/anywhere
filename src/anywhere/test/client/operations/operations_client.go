// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
)

// New creates a new operations API client.
func New(transport runtime.ClientTransport, formats strfmt.Registry) ClientService {
	return &Client{transport: transport, formats: formats}
}

/*
Client for operations API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
}

// ClientService is the interface for Client methods
type ClientService interface {
	GetV1AgentList(params *GetV1AgentListParams) (*GetV1AgentListOK, error)

	GetV1ProxyList(params *GetV1ProxyListParams) (*GetV1ProxyListOK, error)

	GetV1Summary(params *GetV1SummaryParams) (*GetV1SummaryOK, error)

	GetV1SupportIP(params *GetV1SupportIPParams) (*GetV1SupportIPOK, error)

	PostV1ProxyAdd(params *PostV1ProxyAddParams) (*PostV1ProxyAddOK, error)

	PostV1ProxyDelete(params *PostV1ProxyDeleteParams) (*PostV1ProxyDeleteOK, error)

	PostV1ProxyUpdate(params *PostV1ProxyUpdateParams) (*PostV1ProxyUpdateOK, error)

	SetTransport(transport runtime.ClientTransport)
}

/*
  GetV1AgentList returns a list of all agent
*/
func (a *Client) GetV1AgentList(params *GetV1AgentListParams) (*GetV1AgentListOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetV1AgentListParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "GetV1AgentList",
		Method:             "GET",
		PathPattern:        "/v1/agent/list",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/x-www-form-urlencoded"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetV1AgentListReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetV1AgentListOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetV1AgentListDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  GetV1ProxyList returns a list of all proxy config
*/
func (a *Client) GetV1ProxyList(params *GetV1ProxyListParams) (*GetV1ProxyListOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetV1ProxyListParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "GetV1ProxyList",
		Method:             "GET",
		PathPattern:        "/v1/proxy/list",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/x-www-form-urlencoded"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetV1ProxyListReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetV1ProxyListOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetV1ProxyListDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  GetV1Summary get v1 summary API
*/
func (a *Client) GetV1Summary(params *GetV1SummaryParams) (*GetV1SummaryOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetV1SummaryParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "GetV1Summary",
		Method:             "GET",
		PathPattern:        "/v1/summary",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/x-www-form-urlencoded"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetV1SummaryReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetV1SummaryOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetV1SummaryDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  GetV1SupportIP returns this server s public ip
*/
func (a *Client) GetV1SupportIP(params *GetV1SupportIPParams) (*GetV1SupportIPOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetV1SupportIPParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "GetV1SupportIP",
		Method:             "GET",
		PathPattern:        "/v1/support/ip",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/x-www-form-urlencoded"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetV1SupportIPReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetV1SupportIPOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetV1SupportIPDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  PostV1ProxyAdd post v1 proxy add API
*/
func (a *Client) PostV1ProxyAdd(params *PostV1ProxyAddParams) (*PostV1ProxyAddOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewPostV1ProxyAddParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "PostV1ProxyAdd",
		Method:             "POST",
		PathPattern:        "/v1/proxy/add",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/x-www-form-urlencoded"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &PostV1ProxyAddReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*PostV1ProxyAddOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*PostV1ProxyAddDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  PostV1ProxyDelete post v1 proxy delete API
*/
func (a *Client) PostV1ProxyDelete(params *PostV1ProxyDeleteParams) (*PostV1ProxyDeleteOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewPostV1ProxyDeleteParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "PostV1ProxyDelete",
		Method:             "POST",
		PathPattern:        "/v1/proxy/delete",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/x-www-form-urlencoded"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &PostV1ProxyDeleteReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*PostV1ProxyDeleteOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*PostV1ProxyDeleteDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  PostV1ProxyUpdate post v1 proxy update API
*/
func (a *Client) PostV1ProxyUpdate(params *PostV1ProxyUpdateParams) (*PostV1ProxyUpdateOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewPostV1ProxyUpdateParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "PostV1ProxyUpdate",
		Method:             "POST",
		PathPattern:        "/v1/proxy/update",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/x-www-form-urlencoded"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &PostV1ProxyUpdateReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*PostV1ProxyUpdateOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*PostV1ProxyUpdateDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

// SetTransport changes the transport on the client
func (a *Client) SetTransport(transport runtime.ClientTransport) {
	a.transport = transport
}
