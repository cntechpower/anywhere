package http

import (
	"context"

	"github.com/cntechpower/anywhere/server/restapi/api/models"
	pb "github.com/cntechpower/anywhere/server/rpc/definitions"
	"github.com/cntechpower/anywhere/server/rpc/handler"
)

func GetSummaryV1() (*models.SummaryStatistic, error) {
	c, err := handler.NewClient(false)
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
		NetworkFlowTotalCountInBytes: s.ProxyNetFlowInBytes,
		ProxyConfigTotalCount:        s.ProxyCount,
		ProxyConnectRejectCountTop10: make([]*models.ProxyConfig, 0),
		ProxyConnectTotalCount:       s.ProxyConnectCount,
		ProxyNetworkFlowTop10:        make([]*models.ProxyConfig, 0),
	}
	for _, p := range s.ConfigConnectFailTop10 {
		res.ProxyConnectRejectCountTop10 = append(res.ProxyConnectRejectCountTop10, &models.ProxyConfig{
			AgentID:                         p.AgentId,
			IsWhitelistOn:                   p.IsWhiteListOn,
			LocalAddr:                       p.LocalAddr,
			NetworkFlowLocalToRemoteInBytes: p.NetworkFlowLocalToRemoteInBytes,
			NetworkFlowRemoteToLocalInBytes: p.NetworkFlowRemoteToLocalInBytes,
			ProxyConnectCount:               p.ProxyConnectCount,
			ProxyConnectRejectCount:         p.ProxyConnectRejectCount,
			RemotePort:                      p.RemotePort,
			WhitelistIps:                    p.WhiteCidrList,
		})
	}

	for _, p := range s.ConfigNetFlowTop10 {
		res.ProxyNetworkFlowTop10 = append(res.ProxyNetworkFlowTop10, &models.ProxyConfig{
			AgentID:                         p.AgentId,
			IsWhitelistOn:                   p.IsWhiteListOn,
			LocalAddr:                       p.LocalAddr,
			NetworkFlowLocalToRemoteInBytes: p.NetworkFlowLocalToRemoteInBytes,
			NetworkFlowRemoteToLocalInBytes: p.NetworkFlowRemoteToLocalInBytes,
			ProxyConnectCount:               p.ProxyConnectCount,
			ProxyConnectRejectCount:         p.ProxyConnectRejectCount,
			RemotePort:                      p.RemotePort,
			WhitelistIps:                    p.WhiteCidrList,
		})
	}
	return res, nil
}
