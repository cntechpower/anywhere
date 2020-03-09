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

type NetworkConfig struct {
	MainPort    int    `json:"port"`
	GrpcPort    int    `json:"grpc_port"`
	IsWebEnable bool   `json:"is_web_enable"`
	RestAddr    string `json:"rest_api_listen_addr"`
	WebAddr     string `json:"web_ui_listen_addr"`
}

type SslConfig struct {
	CertFile string `json:"cert_file_path"`
	KeyFile  string `json:"key_file_path"`
	CaFile   string `json:"ca_file_path"`
}

type UserConfig struct {
	AdminUser string `json:"admin_user_name"`
	AdminPass string `json:"admin_password"`
}

type SystemConfig struct {
	ServerId string         `json:"server_id"`
	Ssl      *SslConfig     `json:"ssl_config"`
	Net      *NetworkConfig `json:"net_config"`
	User     *UserConfig    `json:"user_config"`
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
