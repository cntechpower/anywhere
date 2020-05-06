package model

import (
	"encoding/json"
	"net"
	"time"
)

type ReqType string

const AnywhereVersion = "0.0.2"

const (
	PkgReqHeartBeatPing    ReqType = "1"
	PkgControlConnRegister ReqType = "2"
	PkgTunnelBegin         ReqType = "3"
	PkgReqHeartBeatPong    ReqType = "4"
)

type AgentRegisterMsg struct {
	AgentId string
}

func NewAgentRegisterMsg(id string) *RequestMsg {
	return newRequestMsg(PkgControlConnRegister, id, "", &AgentRegisterMsg{AgentId: id})

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
		Version: AnywhereVersion,
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

func NewHeartBeatPingMsg(c net.Conn, id string) *RequestMsg {
	return newRequestMsg(PkgReqHeartBeatPing, id, "", &HeartBeatMsg{
		LocalAddr:  c.LocalAddr().String(),
		RemoteAddr: c.RemoteAddr().String(),
		SendTime:   time.Now(),
	})
}

func NewHeartBeatPongMsg(c net.Conn, id string) *RequestMsg {
	return newRequestMsg(PkgReqHeartBeatPong, id, "", &HeartBeatMsg{
		LocalAddr:  c.LocalAddr().String(),
		RemoteAddr: c.RemoteAddr().String(),
		SendTime:   time.Now(),
	})
}

type TunnelBeginMsg struct {
	AgentId   string
	LocalAddr string
}

func NewTunnelBeginMsg(id, addr string) *RequestMsg {
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
