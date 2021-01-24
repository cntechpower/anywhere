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
	"github.com/go-openapi/swag"
)

// NewPostV1ProxyAddParams creates a new PostV1ProxyAddParams object
// with the default values initialized.
func NewPostV1ProxyAddParams() *PostV1ProxyAddParams {
	var (
		whiteListIpsDefault = string("")
	)
	return &PostV1ProxyAddParams{
		WhiteListIps: &whiteListIpsDefault,

		timeout: cr.DefaultTimeout,
	}
}

// NewPostV1ProxyAddParamsWithTimeout creates a new PostV1ProxyAddParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewPostV1ProxyAddParamsWithTimeout(timeout time.Duration) *PostV1ProxyAddParams {
	var (
		whiteListIpsDefault = string("")
	)
	return &PostV1ProxyAddParams{
		WhiteListIps: &whiteListIpsDefault,

		timeout: timeout,
	}
}

// NewPostV1ProxyAddParamsWithContext creates a new PostV1ProxyAddParams object
// with the default values initialized, and the ability to set a context for a request
func NewPostV1ProxyAddParamsWithContext(ctx context.Context) *PostV1ProxyAddParams {
	var (
		whiteListIpsDefault = string("")
	)
	return &PostV1ProxyAddParams{
		WhiteListIps: &whiteListIpsDefault,

		Context: ctx,
	}
}

// NewPostV1ProxyAddParamsWithHTTPClient creates a new PostV1ProxyAddParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewPostV1ProxyAddParamsWithHTTPClient(client *http.Client) *PostV1ProxyAddParams {
	var (
		whiteListIpsDefault = string("")
	)
	return &PostV1ProxyAddParams{
		WhiteListIps: &whiteListIpsDefault,
		HTTPClient:   client,
	}
}

/*PostV1ProxyAddParams contains all the parameters to send to the API endpoint
for the post v1 proxy add operation typically these are written to a http.Request
*/
type PostV1ProxyAddParams struct {

	/*GroupName
	  group name

	*/
	GroupName string
	/*LocalAddr
	  localAddress

	*/
	LocalAddr string
	/*RemotePort
	  anywhered server listen port

	*/
	RemotePort int64
	/*UserName
	  user name

	*/
	UserName string
	/*WhiteListEnable
	  white_list_enable

	*/
	WhiteListEnable bool
	/*WhiteListIps
	  white_list_ips

	*/
	WhiteListIps *string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the post v1 proxy add params
func (o *PostV1ProxyAddParams) WithTimeout(timeout time.Duration) *PostV1ProxyAddParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the post v1 proxy add params
func (o *PostV1ProxyAddParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the post v1 proxy add params
func (o *PostV1ProxyAddParams) WithContext(ctx context.Context) *PostV1ProxyAddParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the post v1 proxy add params
func (o *PostV1ProxyAddParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the post v1 proxy add params
func (o *PostV1ProxyAddParams) WithHTTPClient(client *http.Client) *PostV1ProxyAddParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the post v1 proxy add params
func (o *PostV1ProxyAddParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithGroupName adds the groupName to the post v1 proxy add params
func (o *PostV1ProxyAddParams) WithGroupName(groupName string) *PostV1ProxyAddParams {
	o.SetGroupName(groupName)
	return o
}

// SetGroupName adds the groupName to the post v1 proxy add params
func (o *PostV1ProxyAddParams) SetGroupName(groupName string) {
	o.GroupName = groupName
}

// WithLocalAddr adds the localAddr to the post v1 proxy add params
func (o *PostV1ProxyAddParams) WithLocalAddr(localAddr string) *PostV1ProxyAddParams {
	o.SetLocalAddr(localAddr)
	return o
}

// SetLocalAddr adds the localAddr to the post v1 proxy add params
func (o *PostV1ProxyAddParams) SetLocalAddr(localAddr string) {
	o.LocalAddr = localAddr
}

// WithRemotePort adds the remotePort to the post v1 proxy add params
func (o *PostV1ProxyAddParams) WithRemotePort(remotePort int64) *PostV1ProxyAddParams {
	o.SetRemotePort(remotePort)
	return o
}

// SetRemotePort adds the remotePort to the post v1 proxy add params
func (o *PostV1ProxyAddParams) SetRemotePort(remotePort int64) {
	o.RemotePort = remotePort
}

// WithUserName adds the userName to the post v1 proxy add params
func (o *PostV1ProxyAddParams) WithUserName(userName string) *PostV1ProxyAddParams {
	o.SetUserName(userName)
	return o
}

// SetUserName adds the userName to the post v1 proxy add params
func (o *PostV1ProxyAddParams) SetUserName(userName string) {
	o.UserName = userName
}

// WithWhiteListEnable adds the whiteListEnable to the post v1 proxy add params
func (o *PostV1ProxyAddParams) WithWhiteListEnable(whiteListEnable bool) *PostV1ProxyAddParams {
	o.SetWhiteListEnable(whiteListEnable)
	return o
}

// SetWhiteListEnable adds the whiteListEnable to the post v1 proxy add params
func (o *PostV1ProxyAddParams) SetWhiteListEnable(whiteListEnable bool) {
	o.WhiteListEnable = whiteListEnable
}

// WithWhiteListIps adds the whiteListIps to the post v1 proxy add params
func (o *PostV1ProxyAddParams) WithWhiteListIps(whiteListIps *string) *PostV1ProxyAddParams {
	o.SetWhiteListIps(whiteListIps)
	return o
}

// SetWhiteListIps adds the whiteListIps to the post v1 proxy add params
func (o *PostV1ProxyAddParams) SetWhiteListIps(whiteListIps *string) {
	o.WhiteListIps = whiteListIps
}

// WriteToRequest writes these params to a swagger request
func (o *PostV1ProxyAddParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// form param group_name
	frGroupName := o.GroupName
	fGroupName := frGroupName
	if fGroupName != "" {
		if err := r.SetFormParam("group_name", fGroupName); err != nil {
			return err
		}
	}

	// form param local_addr
	frLocalAddr := o.LocalAddr
	fLocalAddr := frLocalAddr
	if fLocalAddr != "" {
		if err := r.SetFormParam("local_addr", fLocalAddr); err != nil {
			return err
		}
	}

	// form param remote_port
	frRemotePort := o.RemotePort
	fRemotePort := swag.FormatInt64(frRemotePort)
	if fRemotePort != "" {
		if err := r.SetFormParam("remote_port", fRemotePort); err != nil {
			return err
		}
	}

	// form param user_name
	frUserName := o.UserName
	fUserName := frUserName
	if fUserName != "" {
		if err := r.SetFormParam("user_name", fUserName); err != nil {
			return err
		}
	}

	// form param white_list_enable
	frWhiteListEnable := o.WhiteListEnable
	fWhiteListEnable := swag.FormatBool(frWhiteListEnable)
	if fWhiteListEnable != "" {
		if err := r.SetFormParam("white_list_enable", fWhiteListEnable); err != nil {
			return err
		}
	}

	if o.WhiteListIps != nil {

		// form param white_list_ips
		var frWhiteListIps string
		if o.WhiteListIps != nil {
			frWhiteListIps = *o.WhiteListIps
		}
		fWhiteListIps := frWhiteListIps
		if fWhiteListIps != "" {
			if err := r.SetFormParam("white_list_ips", fWhiteListIps); err != nil {
				return err
			}
		}

	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
