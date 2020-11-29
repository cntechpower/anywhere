package model

import (
	"fmt"
	"time"

	"github.com/cntechpower/anywhere/util"
)

type AgentInfoInServer struct {
	Id               string
	UserName         string
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
	AgentId                         string `json:"agent_id"`
	UserName                        string `json:"user_name"`
	RemotePort                      int    `json:"remote_port"`
	LocalAddr                       string `json:"local_addr"`
	IsWhiteListOn                   bool   `json:"is_white_list_enable"`
	WhiteCidrList                   string `json:"white_cidr_list"`
	NetworkFlowRemoteToLocalInBytes uint64 `json:"-"`
	NetworkFlowLocalToRemoteInBytes uint64 `json:"-"`
	ProxyConnectCount               uint64 `json:"-"`
	ProxyConnectRejectCount         uint64 `json:"-"`
}

type UiConfig struct {
	SkipLogin   bool   `json:"skip_login"`
	GrpcAddr    string `json:"grpc_addr"`
	IsWebEnable bool   `json:"is_web_enable"`
	RestAddr    string `json:"rest_api_listen_addr"`
	WebAddr     string `json:"web_ui_listen_addr"`
	DebugMode   bool   `json:"debug"`
}

type SslConfig struct {
	CertFile string `json:"cert_file_path"`
	KeyFile  string `json:"key_file_path"`
	CaFile   string `json:"ca_file_path"`
}

type SystemConfig struct {
	ServerId string      `json:"server_id"`
	MainPort int         `json:"server_port"`
	Ssl      *SslConfig  `json:"ssl_config"`
	UiConfig *UiConfig   `json:"ui_config"`
	User     *UserConfig `json:"user_config"`
}

type ServerSummary struct {
	AgentTotalCount              uint64
	CurrentProxyConnectionCount  uint64
	NetworkFlowTotalCountInBytes uint64
	ProxyConfigTotalCount        uint64
	ProxyConnectRejectCountTop10 []*ProxyConfig
	ProxyConnectTotalCount       uint64
	ProxyConnectRejectCount      uint64
	ProxyNetworkFlowTop10        []*ProxyConfig
	RefreshTime                  time.Time
}

func NewProxyConfig(userName, agentId string, remotePort int, localAddr string, isWhiteListOn bool, whiteListIps string) (*ProxyConfig, error) {

	if err := util.CheckPortValid(remotePort); err != nil {
		return nil, fmt.Errorf("invalid remoteAddr %v in config, error: %v", localAddr, err)
	}
	if err := util.CheckAddrValid(localAddr); err != nil {
		return nil, fmt.Errorf("invalid localAddr %v in config, error: %v", localAddr, err)
	}
	return &ProxyConfig{
		UserName:      userName,
		AgentId:       agentId,
		RemotePort:    remotePort,
		LocalAddr:     localAddr,
		IsWhiteListOn: isWhiteListOn,
		WhiteCidrList: whiteListIps,
	}, nil

}

//TODO: sort
func NewSortedProxyConfigList(list []*ProxyConfig, less func(i, j int) bool) []*ProxyConfig {
	if len(list) <= 1 {
		return list
	}
	res := make([]*ProxyConfig, len(list))
	for _, c := range list {
		res = append(res, c)
	}
	return res
}
