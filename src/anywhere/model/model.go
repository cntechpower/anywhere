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
	RemotePort    int
	LocalAddr     string
	IsWhiteListOn bool
	WhiteListIps  string
}

type GlobalConfig struct {
	ProxyConfigs []*ProxyConfig
}

type SystemConfig struct {
	Port             int
	RestAddress      string
	GrpcPort         int
	ServerId         string
	CertFile         string
	KeyFile          string
	CaFile           string
	IsWebEnable      bool
	WebListenAddress string
	AdminUser        string
	AdminPass        string
}

func NewSystemConfig(port int, restAddr string, grpcPort int, serverId string, certFile, keyFile, caFile string, isWebEnable bool, webAddr, adminUser, adminPass string) *SystemConfig {
	return &SystemConfig{
		Port:             port,
		RestAddress:      restAddr,
		GrpcPort:         grpcPort,
		ServerId:         serverId,
		CertFile:         certFile,
		KeyFile:          keyFile,
		CaFile:           caFile,
		IsWebEnable:      isWebEnable,
		WebListenAddress: webAddr,
		AdminUser:        adminUser,
		AdminPass:        adminPass,
	}
}

func NewProxyConfig(agentId string, remotePort int, localAddr string, isWhiteListOn bool, whiteListIps string) (*ProxyConfig, error) {

	if err := util.CheckPortValid(remotePort); err != nil {
		return nil, fmt.Errorf("invalid remoteAddr %v in config, error: %v", localAddr, err)
	}
	if err := util.CheckAddrValid(localAddr); err != nil {
		return nil, fmt.Errorf("invalid localAddr %v in config, error: %v", localAddr, err)
	}
	return &ProxyConfig{
		AgentId:       agentId,
		RemotePort:    remotePort,
		LocalAddr:     localAddr,
		IsWhiteListOn: isWhiteListOn,
		WhiteListIps:  whiteListIps,
	}, nil

}
