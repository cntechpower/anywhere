// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"anywhere/server/restapi/api/models"
)

// GetV1AgentListOKCode is the HTTP code returned for type GetV1AgentListOK
const GetV1AgentListOKCode int = 200

/*GetV1AgentListOK A JSON array of user names

swagger:response getV1AgentListOK
*/
type GetV1AgentListOK struct {

	/*
	  In: Body
	*/
	Payload []*models.AgentListInfo `json:"body,omitempty"`
}

// NewGetV1AgentListOK creates GetV1AgentListOK with default headers values
func NewGetV1AgentListOK() *GetV1AgentListOK {

	return &GetV1AgentListOK{}
}

// WithPayload adds the payload to the get v1 agent list o k response
func (o *GetV1AgentListOK) WithPayload(payload []*models.AgentListInfo) *GetV1AgentListOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get v1 agent list o k response
func (o *GetV1AgentListOK) SetPayload(payload []*models.AgentListInfo) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetV1AgentListOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if payload == nil {
		// return empty array
		payload = make([]*models.AgentListInfo, 0, 50)
	}

	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}

/*GetV1AgentListDefault generic errors

swagger:response getV1AgentListDefault
*/
type GetV1AgentListDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload models.GenericErrors `json:"body,omitempty"`
}

// NewGetV1AgentListDefault creates GetV1AgentListDefault with default headers values
func NewGetV1AgentListDefault(code int) *GetV1AgentListDefault {
	if code <= 0 {
		code = 500
	}

	return &GetV1AgentListDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the get v1 agent list default response
func (o *GetV1AgentListDefault) WithStatusCode(code int) *GetV1AgentListDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the get v1 agent list default response
func (o *GetV1AgentListDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the get v1 agent list default response
func (o *GetV1AgentListDefault) WithPayload(payload models.GenericErrors) *GetV1AgentListDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get v1 agent list default response
func (o *GetV1AgentListDefault) SetPayload(payload models.GenericErrors) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetV1AgentListDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	payload := o.Payload
	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}
