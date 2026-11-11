package model

import (
	"encoding/json"
	"time"

	"github.com/cntechpower/anywhere/constants"
)

type ReqType string

const (
	PkgReqHeartBeatPing    ReqType = "1"
	PkgControlConnRegister ReqType = "2"
	PkgTunnelBegin         ReqType = "3"
	PkgReqHeartBeatPong    ReqType = "4"
	PkgAuthenticationFail  ReqType = "5"
	PkgUDPData             ReqType = "6"
	PkgUDPDataResp         ReqType = "7"
)

type AgentRegisterMsg struct {
	AgentGroup string
	AgentId    string
	UserName   string
	PassWord   string
	Version    string
}

func NewAgentRegisterMsg(group, id, userName, password string) *RequestMsg {
	return newRequestMsg(PkgControlConnRegister, group, id, "", &AgentRegisterMsg{
		AgentGroup: group,
		AgentId:    id,
		UserName:   userName,
		PassWord:   password,
		Version:    constants.AnywhereVersion,
	})
}

type DataConnRegisterMsg struct {
	AgentId   string
	ProxyAddr string
}

type RequestMsg struct {
	Version   string
	ReqType   ReqType
	FromGroup string
	FromId    string
	To        string
	Message   []byte
}

func newRequestMsg(t ReqType, fromGroup, from, to string, msg interface{}) *RequestMsg {
	j, _ := json.Marshal(msg)
	return &RequestMsg{
		Version:   constants.AnywhereVersion,
		ReqType:   t,
		FromGroup: fromGroup,
		FromId:    from,
		To:        to,
		Message:   j,
	}
}

type ResponseMsg struct {
	Code    int
	Message string
}

func NewResponseMsg(code int, msg string) *ResponseMsg {
	return &ResponseMsg{
		Code:    code,
		Message: msg,
	}
}

type HeartBeatMsg struct {
	LocalAddr  string
	RemoteAddr string
	SendTime   time.Time
}

func NewHeartBeatPingMsg(localAddr, remoteAddr, group, id string) *RequestMsg {
	return newRequestMsg(PkgReqHeartBeatPing, group, id, "", &HeartBeatMsg{
		LocalAddr:  localAddr,
		RemoteAddr: remoteAddr,
		SendTime:   time.Now(),
	})
}

func NewHeartBeatPongMsg(localAddr, remoteAddr, group, id string) *RequestMsg {
	return newRequestMsg(PkgReqHeartBeatPong, group, id, "", &HeartBeatMsg{
		LocalAddr:  localAddr,
		RemoteAddr: remoteAddr,
		SendTime:   time.Now(),
	})
}

type TunnelBeginMsg struct {
	UserName   string
	AgentGroup string
	AgentId    string
	LocalAddr  string
	// Token is a short lived, single use token issued by the server when it
	// asks the agent for a tunnel. It binds a data connection to an
	// authenticated control connection and prevents arbitrary clients from
	// injecting themselves into the proxy connection pool.
	Token string
}

func NewTunnelBeginMsg(userName, group, id, addr, token string) *RequestMsg {
	return newRequestMsg(PkgTunnelBegin, group, id, "", &TunnelBeginMsg{
		UserName:   userName,
		AgentGroup: group,
		AgentId:    id,
		LocalAddr:  addr,
		Token:      token,
	})
}

func ParseHeartBeatPkg(data []byte) (*HeartBeatMsg, error) {
	msg := &HeartBeatMsg{}
	if err := json.Unmarshal(data, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func ParseControlRegisterPkg(data []byte) (*AgentRegisterMsg, error) {
	msg := &AgentRegisterMsg{}
	if err := json.Unmarshal(data, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func ParseTunnelBeginPkg(data []byte) (*TunnelBeginMsg, error) {
	msg := &TunnelBeginMsg{}
	if err := json.Unmarshal(data, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

type AuthenticationFailMsg struct {
	ErrorMsg string
}

func NewAuthenticationFailMsg(errMsg string) *RequestMsg {
	return newRequestMsg(PkgAuthenticationFail, "", "", "", &AuthenticationFailMsg{ErrorMsg: errMsg})
}

func ParseAuthenticationFailMsg(data []byte) (*AuthenticationFailMsg, error) {
	msg := &AuthenticationFailMsg{}
	if err := json.Unmarshal(data, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

// UDPDataMsg carries one UDP datagram in either direction.
type UDPDataMsg struct {
	// SessionID identifies a client endpoint on the server side so replies can
	// be routed back to the right client socket.
	SessionID uint32 `json:"session_id"`
	LocalAddr string `json:"local_addr"`
	Data      []byte `json:"data"`
}

func NewUDPDataMsg(sessionID uint32, localAddr string, data []byte) *RequestMsg {
	body, _ := json.Marshal(&UDPDataMsg{SessionID: sessionID, LocalAddr: localAddr, Data: data})
	return &RequestMsg{
		Version: constants.AnywhereVersion,
		ReqType: PkgUDPData,
		To:      localAddr,
		Message: body,
	}
}

func NewUDPDataRespMsg(sessionID uint32, localAddr string, data []byte) *RequestMsg {
	body, _ := json.Marshal(&UDPDataMsg{SessionID: sessionID, LocalAddr: localAddr, Data: data})
	return &RequestMsg{
		Version: constants.AnywhereVersion,
		ReqType: PkgUDPDataResp,
		To:      localAddr,
		Message: body,
	}
}

func ParseUDPDataMsg(data []byte) (*UDPDataMsg, error) {
	msg := &UDPDataMsg{}
	if err := json.Unmarshal(data, msg); err != nil {
		return nil, err
	}
	return msg, nil
}
