package restHandler

import (
	"anywhere/server/handler/rpcHandler"
	"anywhere/server/restapi/api/models"
	v1 "anywhere/server/restapi/api/restapi/operations"
	pb "anywhere/server/rpc/definitions"
	"anywhere/util"
	"context"
	"net"
)

func ListAgentV1() ([]*models.AgentListInfo, error) {
	c, err := rpcHandler.NewClient()
	if err != nil {
		return nil, err
	}
	res := make([]*models.AgentListInfo, 0)
	agents, err := c.ListAgent(context.Background(), &pb.Empty{})
	if err != nil {
		return nil, err
	}
	for _, agent := range agents.Agent {
		a := &models.AgentListInfo{
			AgentAdminAddr:   agent.AgentRemoteAddr,
			AgentID:          agent.AgentId,
			LastAckSend:      agent.AgentLastAckSend,
			LastAckRcv:       agent.AgentLastAckRcv,
			ProxyConfigCount: agent.AgentProxyConfigCount,
		}
		res = append(res, a)
	}
	return res, nil

}

func ListProxyV1() ([]*models.ProxyConfig, error) {
	c, err := rpcHandler.NewClient()
	if err != nil {
		return nil, err
	}
	res := make([]*models.ProxyConfig, 0)
	configs, err := c.ListProxyConfigs(context.Background(), &pb.Empty{})
	if err != nil {
		return nil, err
	}
	for _, config := range configs.Config {
		res = append(res, &models.ProxyConfig{
			AgentID:                      config.AgentId,
			LocalAddr:                    config.LocalAddr,
			RemotePort:                   config.RemotePort,
			IsWhitelistOn:                config.IsWhiteListOn,
			WhitelistIps:                 config.WhiteCidrList,
			NetworkFlowLocalToRemoteInMb: config.NetworkFlowLocalToRemoteInMB,
			NetworkFlowRemoteToLocalInMb: config.NetworkFlowRemoteToLocalInMB,
		})
	}
	return res, nil
}

func AddProxyConfigV1(params v1.PostV1ProxyAddParams) (*models.ProxyConfig, error) {
	c, err := rpcHandler.NewClient()
	if err != nil {
		return nil, err
	}
	if _, err := c.AddProxyConfig(context.Background(), &pb.AddProxyConfigInput{
		Config: &pb.ProxyConfig{
			AgentId:       params.AgentID,
			RemotePort:    params.RemotePort,
			LocalAddr:     params.LocalAddr,
			IsWhiteListOn: params.WhiteListEnable,
			WhiteCidrList: util.StringNvl(params.WhiteListIps),
		},
	}); err != nil {
		return nil, err
	}
	return &models.ProxyConfig{
		AgentID:       params.AgentID,
		IsWhitelistOn: params.WhiteListEnable,
		LocalAddr:     params.LocalAddr,
		RemotePort:    params.RemotePort,
		WhitelistIps:  util.StringNvl(params.WhiteListIps),
	}, nil

}
func GetV1SupportIP(params v1.GetV1SupportIPParams) (string, error) {
	addr, err := net.ResolveTCPAddr("tcp", params.HTTPRequest.RemoteAddr)
	if err != nil {
		return "", err
	}
	return addr.IP.String(), nil
}

func PostV1ProxyUpdateParams(params v1.PostV1ProxyUpdateParams) (*models.ProxyConfig, error) {
	c, err := rpcHandler.NewClient()
	if err != nil {
		return nil, err
	}
	if _, err := c.UpdateProxyConfigWhiteList(context.Background(), &pb.UpdateProxyConfigWhiteListInput{
		AgentId:         params.AgentID,
		LocalAddr:       params.LocalAddr,
		WhiteCidrs:      util.StringNvl(params.WhiteListIps),
		WhiteListEnable: params.WhiteListEnable,
	}); err != nil {
		return nil, err
	}
	return &models.ProxyConfig{
		AgentID:       params.AgentID,
		IsWhitelistOn: params.WhiteListEnable,
		LocalAddr:     params.LocalAddr,
		WhitelistIps:  util.StringNvl(params.WhiteListIps),
	}, nil
}

func PostV1ProxyDeleteHandler(params v1.PostV1ProxyDeleteParams) (*models.ProxyConfig, error) {
	c, err := rpcHandler.NewClient()
	if err != nil {
		return nil, err
	}
	if _, err := c.RemoveProxyConfig(context.Background(), &pb.RemoveProxyConfigInput{
		AgentId:   params.AgentID,
		LocalAddr: params.LocalAddr,
	}); err != nil {
		return nil, err
	}
	return &models.ProxyConfig{
		AgentID:   params.AgentID,
		LocalAddr: params.LocalAddr,
	}, nil
}

func GetSummaryV1() (*models.SummaryStatistic, error) {
	c, err := rpcHandler.NewClient()
	if err != nil {
		return &models.SummaryStatistic{}, err
	}
	s, err := c.GetSummary(context.Background(), &pb.Empty{})
	if err != nil {
		return &models.SummaryStatistic{}, err
	}

	res := &models.SummaryStatistic{
		AgentTotalCount:              s.AgentCount,
		CurrentProxyConnectionCount:  s.CurrentProxyConnectionCount,
		NetworkFlowTotalCountInMb:    s.ProxyNetFlowInMb,
		ProxyConfigTotalCount:        s.ProxyCount,
		ProxyConnectRejectCountTop10: make([]*models.ProxyConfig, 0),
		ProxyConnectTotalCount:       s.ProxyConnectCount,
		ProxyNetworkFlowTop10:        make([]*models.ProxyConfig, 0),
	}
	for _, p := range s.ConfigConnectFailTop10 {
		res.ProxyConnectRejectCountTop10 = append(res.ProxyConnectRejectCountTop10, &models.ProxyConfig{
			AgentID:                      p.AgentId,
			IsWhitelistOn:                p.IsWhiteListOn,
			LocalAddr:                    p.LocalAddr,
			NetworkFlowLocalToRemoteInMb: p.NetworkFlowLocalToRemoteInMB,
			NetworkFlowRemoteToLocalInMb: p.NetworkFlowRemoteToLocalInMB,
			ProxyConnectCount:            p.ProxyConnectCount,
			ProxyConnectRejectCount:      p.ProxyConnectRejectCount,
			RemotePort:                   p.RemotePort,
			WhitelistIps:                 p.WhiteCidrList,
		})
	}

	for _, p := range s.ConfigNetFlowTop10 {
		res.ProxyNetworkFlowTop10 = append(res.ProxyNetworkFlowTop10, &models.ProxyConfig{
			AgentID:                      p.AgentId,
			IsWhitelistOn:                p.IsWhiteListOn,
			LocalAddr:                    p.LocalAddr,
			NetworkFlowLocalToRemoteInMb: p.NetworkFlowLocalToRemoteInMB,
			NetworkFlowRemoteToLocalInMb: p.NetworkFlowRemoteToLocalInMB,
			ProxyConnectCount:            p.ProxyConnectCount,
			ProxyConnectRejectCount:      p.ProxyConnectRejectCount,
			RemotePort:                   p.RemotePort,
			WhitelistIps:                 p.WhiteCidrList,
		})
	}
	return res, nil
}
