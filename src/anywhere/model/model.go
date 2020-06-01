package model

import (
	"anywhere/util"
	"fmt"
)

type AgentInfoInServer struct {
	Id               string
	RemoteAddr       string
	LastAckSend      string
	LastAckRcv       string
	ProxyConfigCount int
}

type AgentInfoInAgent struct {
	Id          string
	LocalAddr   string
	ServerAddr  string
	LastAckSend string
	LastAckRcv  string
}

type ProxyConfig struct {
	AgentId                 string `json:"agent_id"`
	RemotePort              int    `json:"remote_port"`
	LocalAddr               string `json:"local_addr"`
	IsWhiteListOn           bool   `json:"is_white_list_enable"`
	WhiteCidrList           string `json:"white_cidr_list"`
	NetworkFlowInMb         int    `json:"-"`
	ProxyConnectCount       int    `json:"-"`
	ProxyConnectRejectCount int    `json:"-"`
}

type GlobalConfig struct {
	ProxyConfigs []*ProxyConfig
}

type UiConfig struct {
	SkipLogin   bool   `json:"skip_login"`
	GrpcAddr    string `json:"grpc_addr"`
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
	AdminUser      string `json:"admin_user_name"`
	AdminPass      string `json:"admin_password"`
	AdminOtpEnable bool   `json:"admin_otp_enable"`
	AdminOtpCode   string `json:"admin_otp_code"`
}

type SystemConfig struct {
	ServerId string      `json:"server_id"`
	MainPort int         `json:"server_port"`
	Ssl      *SslConfig  `json:"ssl_config"`
	UiConfig *UiConfig   `json:"ui_config"`
	User     *UserConfig `json:"user_config"`
}

type ServerSummary struct {
	AgentTotalCount              int
	CurrentProxyConnectionCount  int
	NetworkFlowTotalCountInMb    int
	ProxyConfigTotalCount        int
	ProxyConnectRejectCountTop10 []*ProxyConfig
	ProxyConnectTotalCount       int
	ProxyNetworkFlowTop10        []*ProxyConfig
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
		WhiteCidrList: whiteListIps,
	}, nil

}
