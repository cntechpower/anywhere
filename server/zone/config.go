package zone

import (
	"sync/atomic"

	"github.com/cntechpower/anywhere/dao/connlist"

	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/server/api/auth"
)

type ProxyConfigStats struct {
	*model.ProxyConfig
	acl         *auth.WhiteListValidator
	joinedConns *connlist.JoinedConnList
	closeChan   chan struct{}
}

func (c *ProxyConfigStats) AddNetworkFlow(remoteToLocalBytes, localToRemoteBytes uint64) {
	atomic.AddUint64(&c.NetworkFlowLocalToRemoteInBytes, localToRemoteBytes)
	atomic.AddUint64(&c.NetworkFlowRemoteToLocalInBytes, remoteToLocalBytes)
}

func (c *ProxyConfigStats) AddConnectCount(nums uint64) {
	atomic.AddUint64(&c.ProxyConnectCount, nums)
}

func (c *ProxyConfigStats) AddConnectRejectedCount(nums uint64) {
	atomic.AddUint64(&c.ProxyConnectRejectCount, nums)
}

func (c *ProxyConfigStats) GetCurrentConnectionCount() (int64, error) {
	return c.joinedConns.Count()
}
