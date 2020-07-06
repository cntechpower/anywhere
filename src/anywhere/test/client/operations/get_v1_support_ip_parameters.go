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

// NewGetV1SupportIPParams creates a new GetV1SupportIPParams object
// with the default values initialized.
func NewGetV1SupportIPParams() *GetV1SupportIPParams {

	return &GetV1SupportIPParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewGetV1SupportIPParamsWithTimeout creates a new GetV1SupportIPParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewGetV1SupportIPParamsWithTimeout(timeout time.Duration) *GetV1SupportIPParams {

	return &GetV1SupportIPParams{

		timeout: timeout,
	}
}

// NewGetV1SupportIPParamsWithContext creates a new GetV1SupportIPParams object
// with the default values initialized, and the ability to set a context for a request
func NewGetV1SupportIPParamsWithContext(ctx context.Context) *GetV1SupportIPParams {

	return &GetV1SupportIPParams{

		Context: ctx,
	}
}

// NewGetV1SupportIPParamsWithHTTPClient creates a new GetV1SupportIPParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewGetV1SupportIPParamsWithHTTPClient(client *http.Client) *GetV1SupportIPParams {

	return &GetV1SupportIPParams{
		HTTPClient: client,
	}
}

/*GetV1SupportIPParams contains all the parameters to send to the API endpoint
for the get v1 support IP operation typically these are written to a http.Request
*/
type GetV1SupportIPParams struct {
	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the get v1 support IP params
func (o *GetV1SupportIPParams) WithTimeout(timeout time.Duration) *GetV1SupportIPParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get v1 support IP params
func (o *GetV1SupportIPParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get v1 support IP params
func (o *GetV1SupportIPParams) WithContext(ctx context.Context) *GetV1SupportIPParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get v1 support IP params
func (o *GetV1SupportIPParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get v1 support IP params
func (o *GetV1SupportIPParams) WithHTTPClient(client *http.Client) *GetV1SupportIPParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get v1 support IP params
func (o *GetV1SupportIPParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WriteToRequest writes these params to a swagger request
func (o *GetV1SupportIPParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}