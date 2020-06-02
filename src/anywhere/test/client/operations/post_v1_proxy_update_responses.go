// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"anywhere/test/models"
)

// PostV1ProxyUpdateReader is a Reader for the PostV1ProxyUpdate structure.
type PostV1ProxyUpdateReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *PostV1ProxyUpdateReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewPostV1ProxyUpdateOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewPostV1ProxyUpdateDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewPostV1ProxyUpdateOK creates a PostV1ProxyUpdateOK with default headers values
func NewPostV1ProxyUpdateOK() *PostV1ProxyUpdateOK {
	return &PostV1ProxyUpdateOK{}
}

/*PostV1ProxyUpdateOK handles this case with default header values.

A JSON array of user names
*/
type PostV1ProxyUpdateOK struct {
	Payload *models.ProxyConfig
}

func (o *PostV1ProxyUpdateOK) Error() string {
	return fmt.Sprintf("[POST /v1/proxy/update][%d] postV1ProxyUpdateOK  %+v", 200, o.Payload)
}

func (o *PostV1ProxyUpdateOK) GetPayload() *models.ProxyConfig {
	return o.Payload
}

func (o *PostV1ProxyUpdateOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ProxyConfig)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewPostV1ProxyUpdateDefault creates a PostV1ProxyUpdateDefault with default headers values
func NewPostV1ProxyUpdateDefault(code int) *PostV1ProxyUpdateDefault {
	return &PostV1ProxyUpdateDefault{
		_statusCode: code,
	}
}

/*PostV1ProxyUpdateDefault handles this case with default header values.

generic errors
*/
type PostV1ProxyUpdateDefault struct {
	_statusCode int

	Payload models.GenericErrors
}

// Code gets the status code for the post v1 proxy update default response
func (o *PostV1ProxyUpdateDefault) Code() int {
	return o._statusCode
}

func (o *PostV1ProxyUpdateDefault) Error() string {
	return fmt.Sprintf("[POST /v1/proxy/update][%d] PostV1ProxyUpdate default  %+v", o._statusCode, o.Payload)
}

func (o *PostV1ProxyUpdateDefault) GetPayload() models.GenericErrors {
	return o.Payload
}

func (o *PostV1ProxyUpdateDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
