package handler

import (
	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/server/api/http/api/models"
)

// GetSummaryV1 reads the in-process summary cache instead of doing a gRPC
// round trip to ourselves.
func (a *API) GetSummaryV1() (*models.SummaryStatistic, error) {
	s := a.server.GetSummary()
	res := &models.SummaryStatistic{
		AgentTotalCount:              int64(s.AgentTotalCount),
		CurrentProxyConnectionCount:  int64(s.CurrentProxyConnectionCount),
		NetworkFlowTotalCountInBytes: int64(s.NetworkFlowTotalCountInBytes),
		ProxyConfigTotalCount:        int64(s.ProxyConfigTotalCount),
		ProxyConnectRejectCountTop10: make([]*models.ProxyConfig, 0, len(s.ProxyConnectRejectCountTop10)),
		ProxyConnectTotalCount:       int64(s.ProxyConnectTotalCount),
		ProxyConnectRejectCount:      int64(s.ProxyConnectRejectCount),
		ProxyNetworkFlowTop10:        make([]*models.ProxyConfig, 0, len(s.ProxyNetworkFlowTop10)),
	}
	for _, p := range s.ProxyConnectRejectCountTop10 {
		res.ProxyConnectRejectCountTop10 = append(res.ProxyConnectRejectCountTop10, toAPIConfig(p))
	}
	for _, p := range s.ProxyNetworkFlowTop10 {
		res.ProxyNetworkFlowTop10 = append(res.ProxyNetworkFlowTop10, toAPIConfig(p))
	}
	return res, nil
}

func toAPIConfig(p *model.ProxyConfig) *models.ProxyConfig {
	return &models.ProxyConfig{
		ID:                              int64(p.ID),
		UserName:                        p.UserName,
		ZoneName:                        p.ZoneName,
		IsWhitelistOn:                   p.IsWhiteListOn,
		LocalAddr:                       p.LocalAddr,
		NetworkFlowLocalToRemoteInBytes: int64(p.NetworkFlowLocalToRemoteInBytes),
		NetworkFlowRemoteToLocalInBytes: int64(p.NetworkFlowRemoteToLocalInBytes),
		ProxyConnectCount:               int64(p.ProxyConnectCount),
		ProxyConnectRejectCount:         int64(p.ProxyConnectRejectCount),
		RemotePort:                      int64(p.RemotePort),
		WhitelistIps:                    p.WhiteCidrList,
		ListenType:                      p.ListenType,
	}
}
