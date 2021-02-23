package agent

import (
	"sync/atomic"

	"github.com/cntechpower/anywhere/conn"
	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/server/auth"
)

type ProxyConfig struct {
	*model.ProxyConfig
	acl         *auth.WhiteListValidator
	joinedConns *conn.JoinedConnList
	closeChan   chan struct{}
}

func (c *ProxyConfig) AddNetworkFlow(remoteToLocalBytes, localToRemoteBytes uint64) {
	atomic.AddUint64(&c.NetworkFlowLocalToRemoteInBytes, localToRemoteBytes)
	atomic.AddUint64(&c.NetworkFlowRemoteToLocalInBytes, remoteToLocalBytes)
}

func (c *ProxyConfig) AddConnectCount(nums uint64) {
	atomic.AddUint64(&c.ProxyConnectCount, nums)
}

func (c *ProxyConfig) AddConnectRejectedCount(nums uint64) {
	atomic.AddUint64(&c.ProxyConnectRejectCount, nums)
}

func (c *ProxyConfig) GetCurrentConnectionCount() int {
	return c.joinedConns.Count()
}
