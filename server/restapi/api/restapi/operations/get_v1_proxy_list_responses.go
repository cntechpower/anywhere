// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/cntechpower/anywhere/server/restapi/api/models"
)

// GetV1ProxyListOKCode is the HTTP code returned for type GetV1ProxyListOK
const GetV1ProxyListOKCode int = 200

/*GetV1ProxyListOK A JSON array of user names

swagger:response getV1ProxyListOK
*/
type GetV1ProxyListOK struct {

	/*
	  In: Body
	*/
	Payload []*models.ProxyConfig `json:"body,omitempty"`
}

// NewGetV1ProxyListOK creates GetV1ProxyListOK with default headers values
func NewGetV1ProxyListOK() *GetV1ProxyListOK {

	return &GetV1ProxyListOK{}
}

// WithPayload adds the payload to the get v1 proxy list o k response
func (o *GetV1ProxyListOK) WithPayload(payload []*models.ProxyConfig) *GetV1ProxyListOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get v1 proxy list o k response
func (o *GetV1ProxyListOK) SetPayload(payload []*models.ProxyConfig) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetV1ProxyListOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if payload == nil {
		// return empty array
		payload = make([]*models.ProxyConfig, 0, 50)
	}

	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}

/*GetV1ProxyListDefault generic errors

swagger:response getV1ProxyListDefault
*/
type GetV1ProxyListDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload models.GenericErrors `json:"body,omitempty"`
}

// NewGetV1ProxyListDefault creates GetV1ProxyListDefault with default headers values
func NewGetV1ProxyListDefault(code int) *GetV1ProxyListDefault {
	if code <= 0 {
		code = 500
	}

	return &GetV1ProxyListDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the get v1 proxy list default response
func (o *GetV1ProxyListDefault) WithStatusCode(code int) *GetV1ProxyListDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the get v1 proxy list default response
func (o *GetV1ProxyListDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the get v1 proxy list default response
func (o *GetV1ProxyListDefault) WithPayload(payload models.GenericErrors) *GetV1ProxyListDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get v1 proxy list default response
func (o *GetV1ProxyListDefault) SetPayload(payload models.GenericErrors) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetV1ProxyListDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	payload := o.Payload
	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}