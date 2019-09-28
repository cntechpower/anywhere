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
)

type ProxyConfig struct {
	RemoteAddr net.Addr
	LocalAddr  net.Addr
}

func NewProxyConfigMsg(remotePort, localIp, localPort string) *ProxyConfig {
	remoteAddr, err := util.GetAddrByIpPort("0.0.0.0", remotePort)
	if err != nil {
		return nil
	}
	localAddr, err := util.GetAddrByIpPort(localIp, localPort)
	if err != nil {
		return nil
	}
	return &ProxyConfig{
		RemoteAddr: remoteAddr,
		LocalAddr:  localAddr,
	}

}

type RequestMsg struct {
	Version string
	ReqType ReqType
	Message []byte
}

func NewRequestMsg(v string, t ReqType, msg interface{}) *RequestMsg {
	j, _ := json.Marshal(msg)
	return &RequestMsg{
		Version: v,
		ReqType: t,
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
	localAddr  net.Addr
	remoteAddr net.Addr
	sendTime   time.Time
}

func NewHeartBeatMsg(c net.Conn) HeartBeatMsg {
	return HeartBeatMsg{
		localAddr:  c.LocalAddr(),
		remoteAddr: c.RemoteAddr(),
		sendTime:   time.Now(),
	}
}
