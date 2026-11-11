package zone

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/cntechpower/utils/log"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/cntechpower/anywhere/conn"
	"github.com/cntechpower/anywhere/constants"
	"github.com/cntechpower/anywhere/dao/config"
	"github.com/cntechpower/anywhere/dao/connlist"
	"github.com/cntechpower/anywhere/dao/whitelist"
	"github.com/cntechpower/anywhere/internal/trace"
	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/server/api/auth"
	zoneagent "github.com/cntechpower/anywhere/server/zone/agent"
	"github.com/cntechpower/anywhere/util"
)

var errNoHealthyAgent = errors.New("no healthy agent available")

type IZone interface {
	IsAgentExists(agentId string) bool
	RegisterAgent(agentId string, c net.Conn) (isUpdate bool)
	AddProxyConfig(config *model.ProxyConfig) error
	RemoveProxyConfig(remotePort int, localAddr string) error
	ListJoinedConnections() []*model.JoinedConnListItem
	KillJoinedConnById(id uint) error
	FlushJoinedConnections()
	GetCurrentConnectionCount() int
	UpdateProxyConfigWhiteListConfig(remotePort int, localAddr, whiteCidrs string, whiteListEnable bool) error

	Infos() []*model.AgentInfoInServer
	Info() *model.ZoneInfo
	GetProxyConfigCount() int
	ListProxyConfigs() []*model.ProxyConfig

	PutProxyConnection(ctx context.Context, fromAgentId, localAddr, token string, c net.Conn) error
	HandleUDPResponse(msg *model.UDPDataMsg)

	FlushStats()
	CleanupAgents(timeout time.Duration)
	Shutdown()
}

type Zone struct {
	zoneName string
	userName string

	agentsRwMutex sync.RWMutex
	agents        map[string]zoneagent.IAgent

	proxyConfigMutex sync.RWMutex
	proxyConfigs     map[string]*ProxyConfigStats

	connectionPool conn.ConnectionPool
	joinedConns    *connlist.JoinedConnList
	idleTimeout    time.Duration

	closeCh   chan struct{}
	closeOnce sync.Once
	wg        sync.WaitGroup
}

func NewZone(userName, zoneName string, idleTimeout time.Duration) IZone {
	z := &Zone{
		userName:     userName,
		zoneName:     zoneName,
		agents:       make(map[string]zoneagent.IAgent),
		proxyConfigs: make(map[string]*ProxyConfigStats),
		closeCh:      make(chan struct{}),
		joinedConns:  connlist.NewJoinedConnList(userName, zoneName),
		idleTimeout:  idleTimeout,
	}
	z.connectionPool = conn.NewConnectionPool(z.requestNewProxyConn)
	if err := z.restoreProxyConfig(); err != nil {
		log.Errorf(log.NewHeader("restoreProxyConfig"), "restore proxy config for %s/%s: %v", userName, zoneName, err)
	}
	return z
}

func (z *Zone) RegisterAgent(agentId string, c net.Conn) (isUpdate bool) {
	h := log.NewHeader("RegisterAgent")
	z.agentsRwMutex.Lock()
	defer z.agentsRwMutex.Unlock()
	if a, ok := z.agents[agentId]; ok {
		h.Infof("reset admin conn for user %s, zone %s, agent %s", z.userName, z.zoneName, agentId)
		a.ResetAdminConn(c)
		return true
	}
	h.Infof("build admin conn for user %s, zone %s, agent %s", z.userName, z.zoneName, agentId)
	z.agents[agentId] = zoneagent.NewAgentInfo(z.userName, z.zoneName, agentId, c, z.HandleUDPResponse)
	return false
}

func (z *Zone) IsAgentExists(agentId string) bool {
	z.agentsRwMutex.RLock()
	defer z.agentsRwMutex.RUnlock()
	_, ok := z.agents[agentId]
	return ok
}

func (z *Zone) GetCurrentConnectionCount() int {
	return z.joinedConns.Count()
}

func (z *Zone) getProxyConfigMapKey(remotePort int, localAddr string) string {
	return fmt.Sprintf("%d:%s", remotePort, localAddr)
}

// AddProxyConfig binds the listener synchronously so the caller learns about
// port conflicts instead of receiving a success response for a dead config.
func (z *Zone) AddProxyConfig(cfg *model.ProxyConfig) error {
	key := z.getProxyConfigMapKey(cfg.RemotePort, cfg.LocalAddr)
	z.proxyConfigMutex.Lock()
	defer z.proxyConfigMutex.Unlock()
	if _, exist := z.proxyConfigs[key]; exist {
		return fmt.Errorf("proxy config %s already exists in zone %s", key, z.zoneName)
	}

	acl, err := auth.NewWhiteListValidator(cfg.RemotePort, cfg.ZoneName, cfg.LocalAddr, cfg.WhiteCidrList, cfg.IsWhiteListOn)
	if err != nil {
		return fmt.Errorf("init whitelist: %w", err)
	}
	p := newProxyConfigStats(cfg, acl)

	addr := "0.0.0.0:" + strconv.Itoa(cfg.RemotePort)
	if cfg.ListenType == model.ListenTypeUDP {
		ln, err := util.ListenUdp(addr)
		if err != nil {
			return fmt.Errorf("listen udp on port %d: %w", cfg.RemotePort, err)
		}
		p.udp = newUDPProxy(p, ln, func(sessionID uint32, data []byte) error {
			a := z.chooseAgent()
			if a == nil {
				return errNoHealthyAgent
			}
			return a.SendUDPData(sessionID, p.LocalAddr, data)
		})
		z.proxyConfigs[key] = p
		z.wg.Add(1)
		go z.serveUDP(p)
		log.Infof(log.NewHeader("AddProxyConfig"), "added udp proxy %d -> %s in zone %s", cfg.RemotePort, cfg.LocalAddr, z.zoneName)
		return nil
	}

	ln, err := util.ListenTcp(addr)
	if err != nil {
		return fmt.Errorf("listen tcp on port %d: %w", cfg.RemotePort, err)
	}
	p.tcpLn = ln
	z.proxyConfigs[key] = p
	z.wg.Add(1)
	go z.serveTCP(p, ln)
	log.Infof(log.NewHeader("AddProxyConfig"), "added tcp proxy %d -> %s in zone %s", cfg.RemotePort, cfg.LocalAddr, z.zoneName)
	return nil
}

func (z *Zone) RemoveProxyConfig(remotePort int, localAddr string) error {
	key := z.getProxyConfigMapKey(remotePort, localAddr)
	z.proxyConfigMutex.Lock()
	p, ok := z.proxyConfigs[key]
	if ok {
		delete(z.proxyConfigs, key)
	}
	z.proxyConfigMutex.Unlock()
	if !ok {
		return fmt.Errorf("no such proxy config %s", key)
	}
	p.Close()
	return nil
}

func (z *Zone) Infos() []*model.AgentInfoInServer {
	z.agentsRwMutex.RLock()
	defer z.agentsRwMutex.RUnlock()
	res := make([]*model.AgentInfoInServer, 0, len(z.agents))
	for _, a := range z.agents {
		res = append(res, a.Info())
	}
	return res
}

func (z *Zone) Info() *model.ZoneInfo {
	z.agentsRwMutex.RLock()
	count := len(z.agents)
	z.agentsRwMutex.RUnlock()
	return &model.ZoneInfo{
		UserName:    z.userName,
		ZoneName:    z.zoneName,
		AgentsCount: int64(count),
	}
}

func (z *Zone) GetProxyConfigCount() int {
	z.proxyConfigMutex.RLock()
	defer z.proxyConfigMutex.RUnlock()
	return len(z.proxyConfigs)
}

func (z *Zone) ListProxyConfigs() []*model.ProxyConfig {
	z.proxyConfigMutex.RLock()
	defer z.proxyConfigMutex.RUnlock()
	if len(z.proxyConfigs) == 0 {
		return nil
	}
	res := make([]*model.ProxyConfig, 0, len(z.proxyConfigs))
	for _, c := range z.proxyConfigs {
		res = append(res, c.snapshot())
	}
	return res
}

func (z *Zone) UpdateProxyConfigWhiteListConfig(remotePort int, localAddr, whiteCidrs string, whiteListEnable bool) error {
	key := z.getProxyConfigMapKey(remotePort, localAddr)
	z.proxyConfigMutex.RLock()
	c, ok := z.proxyConfigs[key]
	z.proxyConfigMutex.RUnlock()
	if !ok {
		return fmt.Errorf("no such proxy config %s in zone %s", localAddr, z.zoneName)
	}
	// acl is immutable after construction and internally locked.
	c.acl.SetEnable(whiteListEnable)
	if err := c.acl.UpdateCidrs(whiteCidrs); err != nil {
		return err
	}
	c.updateWhiteList(whiteListEnable, whiteCidrs)
	return nil
}

// ---- TCP ----

func (z *Zone) serveTCP(p *ProxyConfigStats, ln *net.TCPListener) {
	defer z.wg.Done()
	h := log.NewHeader(fmt.Sprintf("tunnel-%s-(%d->%s)", p.UserName, p.RemotePort, p.LocalAddr))
	h.Infof("starting new tcp listener on port %d", p.RemotePort)

	sem := make(chan struct{}, constants.ProxyMaxConnectionsPerConfig)
	for {
		c, err := ln.AcceptTCP()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				h.Infof("listener closed")
				return
			}
			select {
			case <-z.closeCh:
				return
			case <-p.closeChan:
				return
			case <-time.After(50 * time.Millisecond):
			}
			h.Errorf("accept new conn error: %v", err)
			continue
		}
		select {
		case <-z.closeCh:
			_ = c.Close()
			return
		case <-p.closeChan:
			_ = c.Close()
			return
		default:
		}
		select {
		case sem <- struct{}{}:
		default:
			p.AddConnectRejectedCount(1)
			_ = c.Close()
			continue
		}
		go func(c *net.TCPConn) {
			defer func() { <-sem }()
			z.handleTCPProxyConnection(c, p)
		}(c)
	}
}

func (z *Zone) handleTCPProxyConnection(c net.Conn, p *ProxyConfigStats) {
	span, ctx := trace.New(context.Background(), "handleTCPProxyConnection")
	defer span.Finish()

	ip, _, err := net.SplitHostPort(c.RemoteAddr().String())
	if err != nil {
		_ = c.Close()
		return
	}
	span.SetTag("remote-ip", ip)
	span.SetTag("local-addr", p.LocalAddr)
	span.SetTag("server-port", p.RemotePort)
	span.SetTag("zone-name", p.ZoneName)

	if !p.acl.IpInWhiteList(ctx, ip) {
		_ = c.Close()
		p.AddConnectRejectedCount(1)
		ext.HTTPStatusCode.Set(span, 403)
		ext.Error.Set(span, true)
		if p.shouldRecordDeny(ip) {
			_ = whitelist.AddWhiteListDenyIp(p.RemotePort, p.UserName, p.ZoneName, p.LocalAddr, ip)
		}
		return
	}

	dst, err := z.connectionPool.Get(ctx, p.LocalAddr)
	if err != nil {
		_ = c.Close()
		p.AddConnectRejectedCount(1)
		log.Infof(log.NewHeader("proxy"), "get conn for %s error: %v", p.LocalAddr, err)
		return
	}
	if dst.GetConn() == nil {
		_ = c.Close()
		p.AddConnectRejectedCount(1)
		return
	}

	idx := z.joinedConns.Add(ctx, conn.NewWrappedConn(p.LocalAddr, c), dst)
	localToRemote, remoteToLocal := conn.JoinConnWithIdleTimeout(dst.GetConn(), c, z.idleTimeout)
	p.AddNetworkFlow(remoteToLocal, localToRemote)
	p.AddConnectCount(1)
	if err := z.joinedConns.Remove(idx); err != nil {
		log.Infof(log.NewHeader("proxy"), "remove joined conn %d: %v", idx, err)
	}
}

// ---- UDP ----

func (z *Zone) serveUDP(p *ProxyConfigStats) {
	defer z.wg.Done()
	h := log.NewHeader(fmt.Sprintf("udp-tunnel-%s-(%d->%s)", p.UserName, p.RemotePort, p.LocalAddr))
	h.Infof("starting new udp listener on port %d", p.RemotePort)

	buf := make([]byte, constants.UDPMaxDatagramSize)
	for {
		n, client, err := p.udp.listener.ReadFromUDP(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			select {
			case <-z.closeCh:
				return
			case <-p.closeChan:
				return
			case <-time.After(50 * time.Millisecond):
			}
			h.Errorf("read udp error: %v", err)
			continue
		}
		ip, _, err := net.SplitHostPort(client.String())
		if err != nil {
			continue
		}
		if !p.acl.IpInWhiteList(nil, ip) {
			p.AddConnectRejectedCount(1)
			if p.shouldRecordDeny(ip) {
				_ = whitelist.AddWhiteListDenyIp(p.RemotePort, p.UserName, p.ZoneName, p.LocalAddr, ip)
			}
			continue
		}
		p.udp.handleDatagram(buf[:n], client)
	}
}

func (z *Zone) HandleUDPResponse(msg *model.UDPDataMsg) {
	if msg == nil {
		return
	}
	z.proxyConfigMutex.RLock()
	var target *ProxyConfigStats
	for _, p := range z.proxyConfigs {
		if p.udp != nil && p.LocalAddr == msg.LocalAddr {
			target = p
			break
		}
	}
	z.proxyConfigMutex.RUnlock()
	if target == nil || target.udp == nil {
		return
	}
	target.udp.handleResponse(msg)
}

// ---- agent selection ----

func (z *Zone) chooseAgent() zoneagent.IAgent {
	z.agentsRwMutex.RLock()
	defer z.agentsRwMutex.RUnlock()
	if len(z.agents) == 0 {
		return nil
	}
	healthy := make([]zoneagent.IAgent, 0, len(z.agents))
	for _, a := range z.agents {
		if a.IsHealthy() {
			healthy = append(healthy, a)
		}
	}
	if len(healthy) == 0 {
		return nil
	}
	//nolint:gosec // load balancing does not need a cryptographic RNG
	return healthy[rand.Intn(len(healthy))]
}

func (z *Zone) requestNewProxyConn(localAddr string) {
	a := z.chooseAgent()
	if a == nil {
		return
	}
	if err := a.AskProxyConn(localAddr); err != nil {
		log.Errorf(log.NewHeader("requestNewProxyConn"), "agent %s request new proxy conn for %s error: %v", a.Info().Id, localAddr, err)
	}
}

func (z *Zone) PutProxyConnection(ctx context.Context, fromAgentId, localAddr, token string, c net.Conn) error {
	z.agentsRwMutex.RLock()
	a, ok := z.agents[fromAgentId]
	z.agentsRwMutex.RUnlock()
	if !ok {
		return fmt.Errorf("agent %s is not registered in zone %s", fromAgentId, z.zoneName)
	}
	if !a.ConsumeProxyToken(localAddr, token) {
		return fmt.Errorf("invalid or expired proxy token for agent %s", fromAgentId)
	}
	return z.connectionPool.Put(ctx, localAddr, conn.NewWrappedConn(fromAgentId, c))
}

// ---- joined conns ----

func (z *Zone) ListJoinedConnections() []*model.JoinedConnListItem {
	return z.joinedConns.List()
}

func (z *Zone) KillJoinedConnById(id uint) error {
	return z.joinedConns.KillById(id)
}

func (z *Zone) FlushJoinedConnections() {
	z.joinedConns.Flush()
}

// ---- lifecycle ----

func (z *Zone) restoreProxyConfig() error {
	header := log.NewHeader(fmt.Sprintf("restoreProxyConfig_%s_%s", z.userName, z.zoneName))
	return config.IteratorByZone(z.userName, z.zoneName, func(c *model.ProxyConfig) {
		if err := z.AddProxyConfig(c); err != nil {
			header.Errorf("restore config %+v error: %v", c, err)
		}
	})
}

// FlushStats persists dirty counters. Called periodically by the server.
func (z *Zone) FlushStats() {
	z.proxyConfigMutex.RLock()
	configs := make([]*ProxyConfigStats, 0, len(z.proxyConfigs))
	for _, p := range z.proxyConfigs {
		configs = append(configs, p)
	}
	z.proxyConfigMutex.RUnlock()
	for _, p := range configs {
		p.flush()
	}
}

// CleanupAgents removes agents that stopped sending heartbeats.
func (z *Zone) CleanupAgents(timeout time.Duration) {
	h := log.NewHeader("CleanupAgents")
	z.agentsRwMutex.Lock()
	var dead []zoneagent.IAgent
	for id, a := range z.agents {
		if a.LastAckRcvTime().Add(timeout).Before(time.Now()) {
			h.Infof("agent %s in zone %s missed heartbeats for %v, removing it", id, z.zoneName, timeout)
			delete(z.agents, id)
			dead = append(dead, a)
		}
	}
	z.agentsRwMutex.Unlock()
	for _, a := range dead {
		a.Close()
	}
}

func (z *Zone) Shutdown() {
	z.closeOnce.Do(func() {
		close(z.closeCh)
		z.proxyConfigMutex.Lock()
		configs := make([]*ProxyConfigStats, 0, len(z.proxyConfigs))
		for key, p := range z.proxyConfigs {
			configs = append(configs, p)
			delete(z.proxyConfigs, key)
		}
		z.proxyConfigMutex.Unlock()
		for _, p := range configs {
			p.flush()
			p.Close()
		}

		z.agentsRwMutex.Lock()
		agents := make([]zoneagent.IAgent, 0, len(z.agents))
		for id, a := range z.agents {
			agents = append(agents, a)
			delete(z.agents, id)
		}
		z.agentsRwMutex.Unlock()
		for _, a := range agents {
			a.Close()
		}

		z.connectionPool.Close()
		z.joinedConns.Flush()
	})
	z.wg.Wait()
}
