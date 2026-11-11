package server

import (
	"context"
	"time"

	"github.com/cntechpower/utils/log"

	"github.com/cntechpower/anywhere/constants"
	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/server/zone"
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

// maintenanceLoop periodically refreshes the summary cache, persists dirty
// proxy counters and removes dead agents.
func (s *Server) maintenanceLoop(ctx context.Context) {
	defer s.wg.Done()
	refreshTicker := time.NewTicker(constants.CacheRefreshLoopTimeSeconds * time.Second)
	cleanupTicker := time.NewTicker(60 * time.Second)
	defer refreshTicker.Stop()
	defer cleanupTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.closeCh:
			return
		case <-refreshTicker.C:
			s.refreshSummary()
			s.flushStats()
		case <-cleanupTicker.C:
			s.cleanupAgents()
		}
	}
}

func (s *Server) snapshotZones() []zone.IZone {
	s.zonesMu.RLock()
	defer s.zonesMu.RUnlock()
	res := make([]zone.IZone, 0, len(s.zones))
	for _, userZones := range s.zones {
		for _, z := range userZones {
			res = append(res, z)
		}
	}
	return res
}

func (s *Server) flushStats() {
	for _, z := range s.snapshotZones() {
		z.FlushStats()
	}
}

func (s *Server) cleanupAgents() {
	timeout := time.Duration(constants.AgentHeartbeatTimeoutSeconds) * time.Second
	for _, z := range s.snapshotZones() {
		z.CleanupAgents(timeout)
	}
}

func (s *Server) refreshSummary() {
	h := log.NewHeader("RefreshSummaryLoop")
	startTime := time.Now()
	newCache := model.ServerSummary{}
	allConfigList := make([]*model.ProxyConfig, 0, 100)

	for _, z := range s.snapshotZones() {
		configs := z.ListProxyConfigs()
		newCache.ProxyConfigTotalCount += uint64(len(configs))
		newCache.AgentTotalCount += uint64(z.Info().AgentsCount)
		newCache.CurrentProxyConnectionCount += uint64(z.GetCurrentConnectionCount())
		allConfigList = append(allConfigList, configs...)
		for _, cfg := range configs {
			newCache.NetworkFlowTotalCountInBytes += cfg.NetworkFlowRemoteToLocalInBytes
			newCache.NetworkFlowTotalCountInBytes += cfg.NetworkFlowLocalToRemoteInBytes
			newCache.ProxyConnectRejectCount += cfg.ProxyConnectRejectCount
			newCache.ProxyConnectTotalCount += cfg.ProxyConnectCount
		}
	}

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
	s.statusRwMutex.Unlock()

	if cost := endTime.Sub(startTime); cost > 500*time.Millisecond {
		log.Infof(h, "refresh summary took %v (%d configs)", cost, len(allConfigList))
	}
}

func (s *Server) GetSummary() model.ServerSummary {
	s.statusRwMutex.RLock()
	res := s.statusCache
	res.ProxyConnectRejectCountTop10 = append([]*model.ProxyConfig(nil), s.statusCache.ProxyConnectRejectCountTop10...)
	res.ProxyNetworkFlowTop10 = append([]*model.ProxyConfig(nil), s.statusCache.ProxyNetworkFlowTop10...)
	s.statusRwMutex.RUnlock()
	return res
}
