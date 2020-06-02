// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"anywhere/server/restapi/api/models"
)

// PostV1ProxyUpdateOKCode is the HTTP code returned for type PostV1ProxyUpdateOK
const PostV1ProxyUpdateOKCode int = 200

/*PostV1ProxyUpdateOK A JSON array of user names

swagger:response postV1ProxyUpdateOK
*/
type PostV1ProxyUpdateOK struct {

	/*
	  In: Body
	*/
	Payload *models.ProxyConfig `json:"body,omitempty"`
}

// NewPostV1ProxyUpdateOK creates PostV1ProxyUpdateOK with default headers values
func NewPostV1ProxyUpdateOK() *PostV1ProxyUpdateOK {

	return &PostV1ProxyUpdateOK{}
}

// WithPayload adds the payload to the post v1 proxy update o k response
func (o *PostV1ProxyUpdateOK) WithPayload(payload *models.ProxyConfig) *PostV1ProxyUpdateOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the post v1 proxy update o k response
func (o *PostV1ProxyUpdateOK) SetPayload(payload *models.ProxyConfig) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PostV1ProxyUpdateOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*PostV1ProxyUpdateDefault generic errors

swagger:response postV1ProxyUpdateDefault
*/
type PostV1ProxyUpdateDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload models.GenericErrors `json:"body,omitempty"`
}

// NewPostV1ProxyUpdateDefault creates PostV1ProxyUpdateDefault with default headers values
func NewPostV1ProxyUpdateDefault(code int) *PostV1ProxyUpdateDefault {
	if code <= 0 {
		code = 500
	}

	return &PostV1ProxyUpdateDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the post v1 proxy update default response
func (o *PostV1ProxyUpdateDefault) WithStatusCode(code int) *PostV1ProxyUpdateDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the post v1 proxy update default response
func (o *PostV1ProxyUpdateDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the post v1 proxy update default response
func (o *PostV1ProxyUpdateDefault) WithPayload(payload models.GenericErrors) *PostV1ProxyUpdateDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the post v1 proxy update default response
func (o *PostV1ProxyUpdateDefault) SetPayload(payload models.GenericErrors) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PostV1ProxyUpdateDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	payload := o.Payload
	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}
