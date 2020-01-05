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

type ProxyConfig struct {
	AgentId       string
	RemoteAddr    string
	LocalAddr     string
	IsWhiteListOn bool
	WhiteListIps  string
}

type GlobalConfig struct {
	ProxyConfigs []ProxyConfig
}

func NewProxyConfig(agentId, remoteAddr string, localAddr string, isWhiteListOn bool, whiteListIps string) (*ProxyConfig, error) {

	if err := util.CheckAddrValid(remoteAddr); err != nil {
		return nil, fmt.Errorf("invalid remoteAddr %v in config, error: %v", localAddr, err)
	}
	if err := util.CheckAddrValid(localAddr); err != nil {
		return nil, fmt.Errorf("invalid localAddr %v in config, error: %v", localAddr, err)
	}
	return &ProxyConfig{
		AgentId:       agentId,
		RemoteAddr:    remoteAddr,
		LocalAddr:     localAddr,
		IsWhiteListOn: isWhiteListOn,
		WhiteListIps:  whiteListIps,
	}, nil

}
