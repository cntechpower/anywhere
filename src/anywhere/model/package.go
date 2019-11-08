package model

import (
	"anywhere/util"
	"encoding/json"
	"net"
	"time"
)

type ReqType string

const (
	PkgReqHeartBeat        ReqType = "ReqHeartBeat"
	PkgControlConnRegister ReqType = "PkgControlConnRegister"
	PkgTunnelBegin         ReqType = "PkgTunnelBegin"
)

type AgentRegisterMsg struct {
	AgentId string
}

func NewAgentRegisterMsg(id string) *AgentRegisterMsg {
	return &AgentRegisterMsg{AgentId: id}
}

type DataConnRegisterMsg struct {
	AgentId   string
	ProxyAddr string
}

type ProxyConfig struct {
	RemoteAddr string
	LocalAddr  string
}

func NewProxyConfigMsg(remotePort int, localIp string, localPort int) (*ProxyConfig, error) {
	remoteAddr, err := util.GetAddrByIpPort("0.0.0.0", remotePort)
	if err != nil {
		return nil, err
	}
	localAddr, err := util.GetAddrByIpPort(localIp, localPort)
	if err != nil {
		return nil, err
	}
	return &ProxyConfig{
		RemoteAddr: remoteAddr.String(),
		LocalAddr:  localAddr.String(),
	}, nil

}

type RequestMsg struct {
	Version string
	ReqType ReqType
	From    string
	To      string
	Message []byte
}

func NewRequestMsg(v string, t ReqType, from, to string, msg interface{}) *RequestMsg {
	j, _ := json.Marshal(msg)
	return &RequestMsg{
		Version: v,
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

func NewHeartBeatMsg(c net.Conn) HeartBeatMsg {
	return HeartBeatMsg{
		LocalAddr:  c.LocalAddr().String(),
		RemoteAddr: c.RemoteAddr().String(),
		SendTime:   time.Now(),
	}
}

type TunnelBeginMsg struct {
	AgentId   string
	LocalAddr string
}

func NewTunnelBeginMsg(id, addr string) *TunnelBeginMsg {
	return &TunnelBeginMsg{AgentId: id, LocalAddr: addr}
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
