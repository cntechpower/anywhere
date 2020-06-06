package anywhereServer

import (
	"anywhere/model"
)

func SortDescAndLimit(a []*model.ProxyConfig, less func(p1 *model.ProxyConfig, p2 *model.ProxyConfig) bool, limit int) []*model.ProxyConfig {
	res := make([]*model.ProxyConfig, 0)
	for _, v := range a {
		if len(res) == 0 {
			res = append(res, v)
			continue
		}

		if len(res) == limit && less(v, res[limit-1]) {
			continue
		}
		inserted := false
		for i, j := range res {
			if !less(v, j) {
				tmp := append(make([]*model.ProxyConfig, 0), res[i:]...)
				res = append(append(res[:i], v), tmp...)
				inserted = true
				break
			}
		}
		if !inserted && len(res) < limit {
			res = append(res, v)
		}
	}
	if len(res) > limit {
		res = res[:limit]
	}
	return res
}

func (s *Server) RefreshSummary() {
	newCache := model.ServerSummary{}
	allConfigList := make([]*model.ProxyConfig, 0)
	for _, agent := range s.agents {
		configs := agent.ListProxyConfigs()
		newCache.ProxyConfigTotalCount += uint64(len(configs))
		allConfigList = append(allConfigList, configs...)
		newCache.AgentTotalCount++
		for _, config := range configs {
			newCache.NetworkFlowTotalCountInBytes += config.NetworkFlowRemoteToLocalInBytes
			newCache.NetworkFlowTotalCountInBytes += config.NetworkFlowLocalToRemoteInBytes
			newCache.ProxyConnectRejectCount += config.ProxyConnectRejectCount
			newCache.ProxyConnectTotalCount += config.ProxyConnectCount
		}
	}
	newCache.ProxyConnectRejectCountTop10 = SortDescAndLimit(allConfigList,
		func(p1 *model.ProxyConfig, p2 *model.ProxyConfig) bool {
			return p1.ProxyConnectRejectCount < p2.ProxyConnectRejectCount
		}, 10)
	newCache.ProxyNetworkFlowTop10 = SortDescAndLimit(allConfigList,
		func(p1 *model.ProxyConfig, p2 *model.ProxyConfig) bool {
			return (p1.NetworkFlowRemoteToLocalInBytes + p1.NetworkFlowLocalToRemoteInBytes) <
				(p2.NetworkFlowRemoteToLocalInBytes + p2.NetworkFlowLocalToRemoteInBytes)
		}, 10)
	s.agentsRwMutex.Lock()
	s.statusCache = newCache
	s.allProxyConfigList = allConfigList
	s.agentsRwMutex.Unlock()
}

func (s *Server) GetSummary() model.ServerSummary {
	return s.statusCache
}
