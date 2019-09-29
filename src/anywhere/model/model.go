package model

import (
	"anywhere/util"
	"encoding/json"
	"net"
	"time"
)

type ReqType string

const (
	PkgReqNewproxy  ReqType = "PkgReqNewproxy"
	PkgReqHeartBeat ReqType = "ReqHeartBeat"
	PkgRegister     ReqType = "PkgRegister"
)

type RegisterMsg struct {
	AgentId string
}

func NewRegisterMsg(id string) *RegisterMsg {
	return &RegisterMsg{AgentId: id}
}

type ProxyConfig struct {
	RemoteAddr string
	LocalAddr  string
}

func NewProxyConfigMsg(remotePort, localIp, localPort string) (*ProxyConfig, error) {
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

func ParseProxyConfig(data []byte) (*ProxyConfig, error) {
	msg := &ProxyConfig{}
	err := json.Unmarshal(data, msg)
	if err != nil {
		return &ProxyConfig{}, err
	}
	return msg, nil
}

func ParseHeartBeatPkg(data []byte) (*HeartBeatMsg, error) {
	msg := &HeartBeatMsg{}
	err := json.Unmarshal(data, msg)
	if err != nil {
		return &HeartBeatMsg{}, err
	}
	return msg, nil

}

func ParseRegisterPkg(data []byte) (*RegisterMsg, error) {
	msg := &RegisterMsg{}
	err := json.Unmarshal(data, msg)
	if err != nil {
		return &RegisterMsg{}, err
	}
	return msg, nil
}
