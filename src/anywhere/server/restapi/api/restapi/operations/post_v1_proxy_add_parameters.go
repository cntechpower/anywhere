// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"

	strfmt "github.com/go-openapi/strfmt"
)

// NewPostV1ProxyAddParams creates a new PostV1ProxyAddParams object
// with the default values initialized.
func NewPostV1ProxyAddParams() PostV1ProxyAddParams {

	var (
		// initialize parameters with default values

		whiteListIpsDefault = string("")
	)

	return PostV1ProxyAddParams{
		WhiteListIps: &whiteListIpsDefault,
	}
}

// PostV1ProxyAddParams contains all the bound params for the post v1 proxy add operation
// typically these are obtained from a http.Request
//
// swagger:parameters PostV1ProxyAdd
type PostV1ProxyAddParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*agent id
	  Required: true
	  In: formData
	*/
	AgentID string
	/*localAddress
	  Required: true
	  In: formData
	*/
	LocalAddr string
	/*anywhered server listen port
	  Required: true
	  In: formData
	*/
	RemotePort int64
	/*white_list_enable
	  Required: true
	  In: formData
	*/
	WhiteListEnable bool
	/*white_list_ips
	  In: formData
	  Default: ""
	*/
	WhiteListIps *string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewPostV1ProxyAddParams() beforehand.
func (o *PostV1ProxyAddParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		if err != http.ErrNotMultipart {
			return errors.New(400, "%v", err)
		} else if err := r.ParseForm(); err != nil {
			return errors.New(400, "%v", err)
		}
	}
	fds := runtime.Values(r.Form)

	fdAgentID, fdhkAgentID, _ := fds.GetOK("agent_id")
	if err := o.bindAgentID(fdAgentID, fdhkAgentID, route.Formats); err != nil {
		res = append(res, err)
	}

	fdLocalAddr, fdhkLocalAddr, _ := fds.GetOK("local_addr")
	if err := o.bindLocalAddr(fdLocalAddr, fdhkLocalAddr, route.Formats); err != nil {
		res = append(res, err)
	}

	fdRemotePort, fdhkRemotePort, _ := fds.GetOK("remote_port")
	if err := o.bindRemotePort(fdRemotePort, fdhkRemotePort, route.Formats); err != nil {
		res = append(res, err)
	}

	fdWhiteListEnable, fdhkWhiteListEnable, _ := fds.GetOK("white_list_enable")
	if err := o.bindWhiteListEnable(fdWhiteListEnable, fdhkWhiteListEnable, route.Formats); err != nil {
		res = append(res, err)
	}

	fdWhiteListIps, fdhkWhiteListIps, _ := fds.GetOK("white_list_ips")
	if err := o.bindWhiteListIps(fdWhiteListIps, fdhkWhiteListIps, route.Formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindAgentID binds and validates parameter AgentID from formData.
func (o *PostV1ProxyAddParams) bindAgentID(rawData []string, hasKey bool, formats strfmt.Registry) error {
	if !hasKey {
		return errors.Required("agent_id", "formData")
	}
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true

	if err := validate.RequiredString("agent_id", "formData", raw); err != nil {
		return err
	}

	o.AgentID = raw

	return nil
}

// bindLocalAddr binds and validates parameter LocalAddr from formData.
func (o *PostV1ProxyAddParams) bindLocalAddr(rawData []string, hasKey bool, formats strfmt.Registry) error {
	if !hasKey {
		return errors.Required("local_addr", "formData")
	}
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true

	if err := validate.RequiredString("local_addr", "formData", raw); err != nil {
		return err
	}

	o.LocalAddr = raw

	return nil
}

// bindRemotePort binds and validates parameter RemotePort from formData.
func (o *PostV1ProxyAddParams) bindRemotePort(rawData []string, hasKey bool, formats strfmt.Registry) error {
	if !hasKey {
		return errors.Required("remote_port", "formData")
	}
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true

	if err := validate.RequiredString("remote_port", "formData", raw); err != nil {
		return err
	}

	value, err := swag.ConvertInt64(raw)
	if err != nil {
		return errors.InvalidType("remote_port", "formData", "int64", raw)
	}
	o.RemotePort = value

	return nil
}

// bindWhiteListEnable binds and validates parameter WhiteListEnable from formData.
func (o *PostV1ProxyAddParams) bindWhiteListEnable(rawData []string, hasKey bool, formats strfmt.Registry) error {
	if !hasKey {
		return errors.Required("white_list_enable", "formData")
	}
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true

	if err := validate.RequiredString("white_list_enable", "formData", raw); err != nil {
		return err
	}

	value, err := swag.ConvertBool(raw)
	if err != nil {
		return errors.InvalidType("white_list_enable", "formData", "bool", raw)
	}
	o.WhiteListEnable = value

	return nil
}

// bindWhiteListIps binds and validates parameter WhiteListIps from formData.
func (o *PostV1ProxyAddParams) bindWhiteListIps(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false

	if raw == "" { // empty values pass all other validations
		// Default values have been previously initialized by NewPostV1ProxyAddParams()
		return nil
	}

	o.WhiteListIps = &raw

	return nil
}
