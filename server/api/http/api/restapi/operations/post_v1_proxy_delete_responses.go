// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/cntechpower/anywhere/server/api/http/api/models"
)

// PostV1ProxyDeleteOKCode is the HTTP code returned for type PostV1ProxyDeleteOK
const PostV1ProxyDeleteOKCode int = 200

/*PostV1ProxyDeleteOK A JSON array of user names

swagger:response postV1ProxyDeleteOK
*/
type PostV1ProxyDeleteOK struct {

	/*
	  In: Body
	*/
	Payload *models.GenericResponse `json:"body,omitempty"`
}

// NewPostV1ProxyDeleteOK creates PostV1ProxyDeleteOK with default headers values
func NewPostV1ProxyDeleteOK() *PostV1ProxyDeleteOK {

	return &PostV1ProxyDeleteOK{}
}

// WithPayload adds the payload to the post v1 proxy delete o k response
func (o *PostV1ProxyDeleteOK) WithPayload(payload *models.GenericResponse) *PostV1ProxyDeleteOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the post v1 proxy delete o k response
func (o *PostV1ProxyDeleteOK) SetPayload(payload *models.GenericResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PostV1ProxyDeleteOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*PostV1ProxyDeleteDefault generic errors

swagger:response postV1ProxyDeleteDefault
*/
type PostV1ProxyDeleteDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload models.GenericErrors `json:"body,omitempty"`
}

// NewPostV1ProxyDeleteDefault creates PostV1ProxyDeleteDefault with default headers values
func NewPostV1ProxyDeleteDefault(code int) *PostV1ProxyDeleteDefault {
	if code <= 0 {
		code = 500
	}

	return &PostV1ProxyDeleteDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the post v1 proxy delete default response
func (o *PostV1ProxyDeleteDefault) WithStatusCode(code int) *PostV1ProxyDeleteDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the post v1 proxy delete default response
func (o *PostV1ProxyDeleteDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the post v1 proxy delete default response
func (o *PostV1ProxyDeleteDefault) WithPayload(payload models.GenericErrors) *PostV1ProxyDeleteDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the post v1 proxy delete default response
func (o *PostV1ProxyDeleteDefault) SetPayload(payload models.GenericErrors) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PostV1ProxyDeleteDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	payload := o.Payload
	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}
