package server

import (
	"context"
	"time"

	"github.com/cntechpower/utils/log"

	"github.com/cntechpower/anywhere/constants"
	"github.com/cntechpower/anywhere/model"
)

func SortDescAndLimitUsingHeap(a []*model.ProxyConfig, less func(p1 *model.ProxyConfig, p2 *model.ProxyConfig) bool, limit int) []*model.ProxyConfig {
	res := make([]*model.ProxyConfig, 0, limit)
	ph := model.InitProxyConfigHeap(a, less, limit)
	for !ph.IsEmpty() {
		res = append(res, ph.Pop())
	}
	return res
}

func SortDescAndLimit(a []*model.ProxyConfig, less func(p1 *model.ProxyConfig, p2 *model.ProxyConfig) bool, limit int) []*model.ProxyConfig {
	res := make([]*model.ProxyConfig, 0, limit+1)
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
		if len(res) > limit {
			res = res[:limit]
		}
	}

	return res
}

func (s *Server) RefreshSummaryLoop(ctx context.Context) {
	h := log.NewHeader("RefreshSummaryLoop")
	currentLoop := 0
	shouldLog := func() bool {
		if currentLoop == constants.CacheRefreshLogInhibition {
			currentLoop = 0
			return true
		}
		currentLoop++
		return false
	}
	ticker := time.NewTicker(constants.CacheRefreshLoopTimeSeconds * time.Second)

	for {
		select {
		case <-ctx.Done():
			h.Infoc(ctx, "existing")
			return
		case <-ticker.C:
		}

		startTime := time.Now()
		newCache := model.ServerSummary{}
		allConfigList := make([]*model.ProxyConfig, 0, 100)
		s.agentsRwMutex.Lock()
		for _, groups := range s.zones {
			for _, group := range groups {
				configs := group.ListProxyConfigs()
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
		}
		s.agentsRwMutex.Unlock()
		newCache.ProxyConnectRejectCountTop10 = SortDescAndLimitUsingHeap(allConfigList,
			func(p1 *model.ProxyConfig, p2 *model.ProxyConfig) bool {
				return p1.ProxyConnectRejectCount > p2.ProxyConnectRejectCount
			}, 10)
		newCache.ProxyNetworkFlowTop10 = SortDescAndLimitUsingHeap(allConfigList,
			func(p1 *model.ProxyConfig, p2 *model.ProxyConfig) bool {
				return (p1.NetworkFlowRemoteToLocalInBytes + p1.NetworkFlowLocalToRemoteInBytes) >
					(p2.NetworkFlowRemoteToLocalInBytes + p2.NetworkFlowLocalToRemoteInBytes)
			}, 10)
		endTime := time.Now()
		newCache.RefreshTime = endTime
		s.statusRwMutex.Lock()
		s.statusCache = newCache
		s.allProxyConfigList = allConfigList
		s.statusRwMutex.Unlock()
		if shouldLog() {
			log.Infof(h, "refresh done, microseconds used %+v", endTime.Sub(startTime).Microseconds())
		}
	}
}

func (s *Server) GetSummary() model.ServerSummary {
	s.statusRwMutex.RLock()
	defer s.statusRwMutex.RUnlock()
	return s.statusCache
}
