package handler

import (
	"net/http"

	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/server/api/http/api/models"
	srv "github.com/cntechpower/anywhere/server/server"
)

// GetSummaryV1 reads the in-process summary cache instead of doing a gRPC
// round trip to ourselves.
func (a *API) GetSummaryV1(r *http.Request) (*models.SummaryStatistic, error) {
	if scoped, restricted := a.scopedUser(r); restricted {
		return a.scopedSummary(scoped)
	}
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

func (a *API) scopedSummary(userName string) (*models.SummaryStatistic, error) {
	res := &models.SummaryStatistic{
		ProxyConnectRejectCountTop10: make([]*models.ProxyConfig, 0),
		ProxyNetworkFlowTop10:        make([]*models.ProxyConfig, 0),
	}
	for _, z := range a.server.ListZones() {
		if z.UserName == userName {
			res.AgentTotalCount += z.AgentsCount
		}
	}
	if groups, err := a.server.ListJoinedConns(userName, ""); err == nil {
		for _, g := range groups {
			res.CurrentProxyConnectionCount += int64(len(g.List))
		}
	}
	userConfigs := make([]*model.ProxyConfig, 0)
	for _, cfg := range a.server.ListProxyConfigs() {
		if cfg.UserName != userName {
			continue
		}
		userConfigs = append(userConfigs, cfg)
		res.ProxyConfigTotalCount++
		res.NetworkFlowTotalCountInBytes += int64(cfg.NetworkFlowRemoteToLocalInBytes + cfg.NetworkFlowLocalToRemoteInBytes)
		res.ProxyConnectTotalCount += int64(cfg.ProxyConnectCount)
		res.ProxyConnectRejectCount += int64(cfg.ProxyConnectRejectCount)
	}
	for _, p := range srv.SortDescAndLimitUsingHeap(userConfigs, func(p1, p2 *model.ProxyConfig) bool {
		return p1.ProxyConnectRejectCount < p2.ProxyConnectRejectCount
	}, 10) {
		res.ProxyConnectRejectCountTop10 = append(res.ProxyConnectRejectCountTop10, toAPIConfig(p))
	}
	for _, p := range srv.SortDescAndLimitUsingHeap(userConfigs, func(p1, p2 *model.ProxyConfig) bool {
		return (p1.NetworkFlowRemoteToLocalInBytes + p1.NetworkFlowLocalToRemoteInBytes) <
			(p2.NetworkFlowRemoteToLocalInBytes + p2.NetworkFlowLocalToRemoteInBytes)
	}, 10) {
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
