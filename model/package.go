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
)

type AgentRegisterMsg struct {
	AgentGroup string
	AgentId    string
	UserName   string
	PassWord   string
}

func NewAgentRegisterMsg(group, id, userName, password string) *RequestMsg {
	return newRequestMsg(PkgControlConnRegister, group, id, "", &AgentRegisterMsg{
		AgentGroup: group,
		AgentId:    id,
		UserName:   userName,
		PassWord:   password,
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
}

func NewTunnelBeginMsg(userName, group, id, addr string) *RequestMsg {
	return newRequestMsg(PkgTunnelBegin, group, id, "", &TunnelBeginMsg{UserName: userName, AgentGroup: group, AgentId: id, LocalAddr: addr})
}

func ParseHeartBeatPkg(data []byte) (*HeartBeatMsg, error) {
	msg := &HeartBeatMsg{}
	err := json.Unmarshal(data, msg)
	if err != nil {
		return &HeartBeatMsg{}, err
	}
	return msg, nil

}

func ParseControlRegisterPkg(data []byte) (*AgentRegisterMsg, error) {
	msg := &AgentRegisterMsg{}
	err := json.Unmarshal(data, msg)
	if err != nil {
		return &AgentRegisterMsg{}, err
	}
	return msg, nil
}

func ParseTunnelBeginPkg(data []byte) (*TunnelBeginMsg, error) {
	msg := &TunnelBeginMsg{}
	err := json.Unmarshal(data, msg)
	if err != nil {
		return &TunnelBeginMsg{}, err
	}
	return msg, nil
}

type AuthenticationFailMsg struct {
	errorMsg string
}

func NewAuthenticationFailMsg(errMsg string) *RequestMsg {
	return newRequestMsg(PkgAuthenticationFail, "", "", "", &AuthenticationFailMsg{
		errorMsg: errMsg,
	})

}

func ParseAuthenticationFailMsg(data []byte) (*AuthenticationFailMsg, error) {
	msg := &AuthenticationFailMsg{}
	err := json.Unmarshal(data, msg)
	if err != nil {
		return &AuthenticationFailMsg{}, err
	}
	return msg, nil
}
