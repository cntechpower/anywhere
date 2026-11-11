package zone

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cntechpower/anywhere/dao/config"
	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/server/api/auth"
)

// ProxyConfigStats embeds the persisted proxy config and keeps the hot path
// counters in dedicated atomics. Counters are flushed to the database
// periodically by Zone.FlushStats instead of on every connection, which used to
// turn into a write storm under load.
//
// The embedded model is treated as immutable after construction (id included:
// the row is inserted before the config is published) except for the whitelist
// fields, which are guarded by cfgMu.
type ProxyConfigStats struct {
	*model.ProxyConfig

	cfgMu sync.RWMutex

	flowLocalToRemote atomic.Uint64
	flowRemoteToLocal atomic.Uint64
	connectCount      atomic.Uint64
	rejectCount       atomic.Uint64

	acl       *auth.WhiteListValidator
	closeChan chan struct{}
	closeOnce sync.Once
	dirty     atomic.Bool
	// tcpLn is the TCP listener of this config (nil for UDP configs).
	tcpLn *net.TCPListener
	// udp holds the sessions of a UDP proxy config; nil for TCP configs.
	udp *udpProxy

	denyMu sync.Mutex
	denyAt map[string]time.Time
}

func newProxyConfigStats(cfg *model.ProxyConfig, acl *auth.WhiteListValidator) *ProxyConfigStats {
	return &ProxyConfigStats{
		ProxyConfig: cfg,
		acl:         acl,
		closeChan:   make(chan struct{}),
	}
}

// Close stops the listener and the UDP session manager. Closing the listener
// is what unblocks the serving goroutine; only closing closeChan leaves
// Accept() blocked forever.
func (c *ProxyConfigStats) Close() {
	c.closeOnce.Do(func() {
		close(c.closeChan)
		if c.tcpLn != nil {
			_ = c.tcpLn.Close()
		}
		if c.udp != nil {
			c.udp.Close()
		}
	})
}

func (c *ProxyConfigStats) closed() bool {
	select {
	case <-c.closeChan:
		return true
	default:
		return false
	}
}

// shouldRecordDeny rate limits whitelist deny records to one per source IP per
// minute so a scanner cannot turn every rejected packet into a database write.
func (c *ProxyConfigStats) shouldRecordDeny(ip string) bool {
	c.denyMu.Lock()
	defer c.denyMu.Unlock()
	if c.denyAt == nil {
		c.denyAt = make(map[string]time.Time)
	}
	if last, ok := c.denyAt[ip]; ok && time.Since(last) < time.Minute {
		return false
	}
	if len(c.denyAt) > 4096 {
		c.denyAt = make(map[string]time.Time)
	}
	c.denyAt[ip] = time.Now()
	return true
}

func (c *ProxyConfigStats) AddNetworkFlow(remoteToLocalBytes, localToRemoteBytes uint64) {
	c.flowLocalToRemote.Add(localToRemoteBytes)
	c.flowRemoteToLocal.Add(remoteToLocalBytes)
	c.dirty.Store(true)
}

func (c *ProxyConfigStats) AddConnectCount(nums uint64) {
	c.connectCount.Add(nums)
	c.dirty.Store(true)
}

func (c *ProxyConfigStats) AddConnectRejectedCount(nums uint64) {
	c.rejectCount.Add(nums)
	c.dirty.Store(true)
}

// snapshot returns a consistent copy of the config with atomically loaded
// counters, safe to hand to the database layer.
func (c *ProxyConfigStats) snapshot() *model.ProxyConfig {
	c.cfgMu.RLock()
	isWhiteListOn := c.IsWhiteListOn
	whiteCidrList := c.WhiteCidrList
	c.cfgMu.RUnlock()

	return &model.ProxyConfig{
		Model:                           c.ProxyConfig.Model,
		UserName:                        c.UserName,
		ZoneName:                        c.ZoneName,
		RemotePort:                      c.RemotePort,
		LocalAddr:                       c.LocalAddr,
		IsWhiteListOn:                   isWhiteListOn,
		WhiteCidrList:                   whiteCidrList,
		NetworkFlowRemoteToLocalInBytes: c.flowRemoteToLocal.Load(),
		NetworkFlowLocalToRemoteInBytes: c.flowLocalToRemote.Load(),
		ProxyConnectCount:               c.connectCount.Load(),
		ProxyConnectRejectCount:         c.rejectCount.Load(),
		ListenType:                      c.ListenType,
	}
}

// updateWhiteList stores the new whitelist settings under the config lock.
func (c *ProxyConfigStats) updateWhiteList(enable bool, cidrs string) {
	c.cfgMu.Lock()
	c.IsWhiteListOn = enable
	c.WhiteCidrList = cidrs
	c.cfgMu.Unlock()
	c.dirty.Store(true)
}

// flush persists the counters if they changed since the last flush.
func (c *ProxyConfigStats) flush() {
	if !c.dirty.Swap(false) {
		return
	}
	if err := config.Save(c.snapshot()); err != nil {
		c.dirty.Store(true)
	}
}
