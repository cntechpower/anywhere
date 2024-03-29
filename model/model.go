package model

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/cntechpower/anywhere/util"
)

type AgentInfoInServer struct {
	Id               string
	ZoneName         string
	UserName         string
	RemoteAddr       string
	LastAckSend      time.Time
	LastAckRcv       time.Time
	ProxyConfigCount int
}

type ZoneInfo struct {
	UserName    string
	ZoneName    string
	AgentsCount int64
}

type AgentInfoInAgent struct {
	Id          string
	LocalAddr   string
	ServerAddr  string
	LastAckSend string
	LastAckRcv  string
}

type ProxyConfigs struct {
	ProxyConfigs map[string] /*user*/ []*ProxyConfig
}

const (
	ListenTypeTCP = "tcp"
	ListenTypeUDP = "udp"
)

type ProxyConfig struct {
	gorm.Model
	UserName                        string `json:"user_name"`
	ZoneName                        string `json:"zone_name"`
	RemotePort                      int    `json:"remote_port"`
	LocalAddr                       string `json:"local_addr"`
	IsWhiteListOn                   bool   `json:"is_white_list_enable"`
	WhiteCidrList                   string `json:"white_cidr_list"`
	NetworkFlowRemoteToLocalInBytes uint64 `json:"network_in_bytes"`
	NetworkFlowLocalToRemoteInBytes uint64 `json:"network_out_bytes"`
	ProxyConnectCount               uint64 `json:"proxy_connect_count"`
	ProxyConnectRejectCount         uint64 `json:"proxy_connect_reject_count"`
	//监听类型: TCP/UDP
	ListenType string `json:"listen_type"`
}

type WhiteListDenyRecord struct {
	gorm.Model
	UserName   string `json:"user_name"`
	ZoneName   string `json:"zone_name"`
	RemotePort int    `json:"remote_port"`
	LocalAddr  string `json:"local_addr"`
	IP         string `json:"ip"`
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

type JoinedConnListItem struct {
	gorm.Model
	UserName      string
	ZoneName      string
	SrcName       string
	DstName       string
	SrcRemoteAddr string
	SrcLocalAddr  string
	DstRemoteAddr string
	DstLocalAddr  string
}

type GroupConnList struct {
	UserName string
	ZoneName string
	List     []*JoinedConnListItem
}

func NewProxyConfig(userName, zoneName string, remotePort int, localAddr string, isWhiteListOn bool, whiteListIps, listenType string) (*ProxyConfig, error) {
	if err := util.CheckPortValid(remotePort); err != nil {
		return nil, fmt.Errorf("invalid remoteAddr %v in config, error: %v", localAddr, err)
	}
	if err := util.CheckAddrValid(localAddr); err != nil {
		return nil, fmt.Errorf("invalid localAddr %v in config, error: %v", localAddr, err)
	}
	return &ProxyConfig{
		UserName:      userName,
		ZoneName:      zoneName,
		RemotePort:    remotePort,
		LocalAddr:     localAddr,
		IsWhiteListOn: isWhiteListOn,
		WhiteCidrList: whiteListIps,
		ListenType:    listenType,
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

func GetPersistModels() []interface{} {
	res := make([]interface{}, 0)
	res = append(res,
		&ProxyConfig{},
		&WhiteListDenyRecord{},
	)
	return res
}

func GetTmpModels() []interface{} {
	res := make([]interface{}, 0)
	res = append(res,
		&JoinedConnListItem{},
	)
	return res
}
