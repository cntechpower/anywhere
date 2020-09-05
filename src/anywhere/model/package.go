package model

import (
	"anywhere/constants"
	"encoding/json"
	"time"
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
	AgentId  string
	UserName string
	PassWord string
}

func NewAgentRegisterMsg(id, userName, password string) *RequestMsg {
	return newRequestMsg(PkgControlConnRegister, id, "", &AgentRegisterMsg{
		AgentId:  id,
		UserName: userName,
		PassWord: password,
	})
}

type DataConnRegisterMsg struct {
	AgentId   string
	ProxyAddr string
}

type RequestMsg struct {
	Version string
	ReqType ReqType
	From    string
	To      string
	Message []byte
}

func newRequestMsg(t ReqType, from, to string, msg interface{}) *RequestMsg {
	j, _ := json.Marshal(msg)
	return &RequestMsg{
		Version: constants.AnywhereVersion,
		ReqType: t,
		From:    from,
		To:      to,
		Message: j,
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

func NewHeartBeatPingMsg(localAddr, remoteAddr, id string) *RequestMsg {
	return newRequestMsg(PkgReqHeartBeatPing, id, "", &HeartBeatMsg{
		LocalAddr:  localAddr,
		RemoteAddr: remoteAddr,
		SendTime:   time.Now(),
	})
}

func NewHeartBeatPongMsg(localAddr, remoteAddr, id string) *RequestMsg {
	return newRequestMsg(PkgReqHeartBeatPong, id, "", &HeartBeatMsg{
		LocalAddr:  localAddr,
		RemoteAddr: remoteAddr,
		SendTime:   time.Now(),
	})
}

type TunnelBeginMsg struct {
	UserName  string
	AgentId   string
	LocalAddr string
}

func NewTunnelBeginMsg(userName, id, addr string) *RequestMsg {
	return newRequestMsg(PkgTunnelBegin, id, "", &TunnelBeginMsg{AgentId: id, LocalAddr: addr})
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
	return newRequestMsg(PkgAuthenticationFail, "", "", &AuthenticationFailMsg{
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
