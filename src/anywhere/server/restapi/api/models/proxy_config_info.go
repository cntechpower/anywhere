// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/swag"
)

// ProxyConfigInfo proxy config information
// swagger:model ProxyConfigInfo
type ProxyConfigInfo struct {

	// agent id
	AgentID string `json:"agent_id,omitempty"`

	// is whitelist on
	IsWhitelistOn bool `json:"is_whitelist_on,omitempty"`

	// localAddress
	LocalAddr string `json:"local_addr,omitempty"`

	// anywhered server listen addr
	RemoteAddr string `json:"remote_addr,omitempty"`

	// whitelist ips
	WhitelistIps string `json:"whitelist_ips,omitempty"`
}

// Validate validates this proxy config info
func (m *ProxyConfigInfo) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *ProxyConfigInfo) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ProxyConfigInfo) UnmarshalBinary(b []byte) error {
	var res ProxyConfigInfo
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
