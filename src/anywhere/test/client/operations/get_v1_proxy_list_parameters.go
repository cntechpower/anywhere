// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NewGetV1ProxyListParams creates a new GetV1ProxyListParams object
// with the default values initialized.
func NewGetV1ProxyListParams() *GetV1ProxyListParams {

	return &GetV1ProxyListParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewGetV1ProxyListParamsWithTimeout creates a new GetV1ProxyListParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewGetV1ProxyListParamsWithTimeout(timeout time.Duration) *GetV1ProxyListParams {

	return &GetV1ProxyListParams{

		timeout: timeout,
	}
}

// NewGetV1ProxyListParamsWithContext creates a new GetV1ProxyListParams object
// with the default values initialized, and the ability to set a context for a request
func NewGetV1ProxyListParamsWithContext(ctx context.Context) *GetV1ProxyListParams {

	return &GetV1ProxyListParams{

		Context: ctx,
	}
}

// NewGetV1ProxyListParamsWithHTTPClient creates a new GetV1ProxyListParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewGetV1ProxyListParamsWithHTTPClient(client *http.Client) *GetV1ProxyListParams {

	return &GetV1ProxyListParams{
		HTTPClient: client,
	}
}

/*GetV1ProxyListParams contains all the parameters to send to the API endpoint
for the get v1 proxy list operation typically these are written to a http.Request
*/
type GetV1ProxyListParams struct {
	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the get v1 proxy list params
func (o *GetV1ProxyListParams) WithTimeout(timeout time.Duration) *GetV1ProxyListParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get v1 proxy list params
func (o *GetV1ProxyListParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get v1 proxy list params
func (o *GetV1ProxyListParams) WithContext(ctx context.Context) *GetV1ProxyListParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get v1 proxy list params
func (o *GetV1ProxyListParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get v1 proxy list params
func (o *GetV1ProxyListParams) WithHTTPClient(client *http.Client) *GetV1ProxyListParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get v1 proxy list params
func (o *GetV1ProxyListParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WriteToRequest writes these params to a swagger request
func (o *GetV1ProxyListParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}