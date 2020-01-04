package model

import (
	"anywhere/util"
	"fmt"
)

type AgentInfo struct {
	Id          string
	RemoteAddr  string
	LastAckSend string
	LastAckRcv  string
}

type ProxyConfigInfo struct {
	AgentId    string
	RemoteAddr string
	LocalAddr  string
}

type ProxyConfig struct {
	AgentId       string
	RemoteAddr    string
	LocalAddr     string
	IsWhiteListOn bool
	WhiteListIps  string
}

func NewProxyConfig(remotePort int, localAddr string, isWhiteListOn bool, whiteListIps string) (*ProxyConfig, error) {
	remoteAddr, err := util.GetAddrByIpPort("0.0.0.0", remotePort)
	if err != nil {
		return nil, err
	}
	if err := util.CheckAddrValid(localAddr); err != nil {
		return nil, fmt.Errorf("invalid localAddr %v in config, error: %v", localAddr, err)
	}
	return &ProxyConfig{
		RemoteAddr:    remoteAddr.String(),
		LocalAddr:     localAddr,
		IsWhiteListOn: isWhiteListOn,
		WhiteListIps:  whiteListIps,
	}, nil

}
