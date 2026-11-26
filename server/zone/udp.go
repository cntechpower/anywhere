package zone

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cntechpower/utils/log"

	"github.com/cntechpower/anywhere/constants"
	"github.com/cntechpower/anywhere/model"
)

// nextUDPSessionID allocates process-wide unique UDP session IDs so multiple
// UDP proxy configs in the same zone never collide on the agent side.
var nextUDPSessionID atomic.Uint32

type udpOutbound struct {
	session *udpSession
	payload []byte
}

type udpSession struct {
	id         uint32
	clientAddr *net.UDPAddr
	lastActive atomic.Int64 // unix nano

	agentMu sync.Mutex
	agentID string
}

func (s *udpSession) touch() { s.lastActive.Store(time.Now().UnixNano()) }

func (s *udpSession) idleSince() time.Duration {
	return time.Since(time.Unix(0, s.lastActive.Load()))
}

func (s *udpSession) getAgentID() string {
	s.agentMu.Lock()
	defer s.agentMu.Unlock()
	return s.agentID
}

func (s *udpSession) setAgentID(id string) {
	s.agentMu.Lock()
	s.agentID = id
	s.agentMu.Unlock()
}

// udpProxy keeps the state of one UDP listener: which client endpoint maps to
// which session ID and how to send datagrams towards the agent.
type udpProxy struct {
	cfg      *ProxyConfigStats
	listener *net.UDPConn
	sendFn   func(preferredAgentID string, sessionID uint32, data []byte) (string, error)
	h        *log.Header

	mu     sync.Mutex
	byID   map[uint32]*udpSession
	byAddr map[string]*udpSession

	queue     chan udpOutbound
	done      chan struct{}
	closeOnce sync.Once
}

func newUDPProxy(cfg *ProxyConfigStats, listener *net.UDPConn, sendFn func(preferredAgentID string, sessionID uint32, data []byte) (string, error)) *udpProxy {
	p := &udpProxy{
		cfg:      cfg,
		listener: listener,
		sendFn:   sendFn,
		h:        log.NewHeader("udp:" + cfg.LocalAddr),
		byID:     make(map[uint32]*udpSession),
		byAddr:   make(map[string]*udpSession),
		queue:    make(chan udpOutbound, constants.UDPDataChannelBuffer),
		done:     make(chan struct{}),
	}
	go p.writeLoop()
	go p.cleanupLoop()
	return p
}

func (p *udpProxy) session(client *net.UDPAddr) *udpSession {
	key := client.String()
	p.mu.Lock()
	defer p.mu.Unlock()
	if s, ok := p.byAddr[key]; ok {
		s.touch()
		return s
	}
	id := nextUDPSessionID.Add(1)
	if id == 0 {
		id = nextUDPSessionID.Add(1)
	}
	s := &udpSession{id: id, clientAddr: client}
	s.touch()
	p.byID[id] = s
	p.byAddr[key] = s
	p.cfg.AddConnectCount(1)
	return s
}

func (p *udpProxy) hasSession(id uint32) bool {
	p.mu.Lock()
	_, ok := p.byID[id]
	p.mu.Unlock()
	return ok
}

func (p *udpProxy) handleDatagram(data []byte, client *net.UDPAddr) {
	s := p.session(client)
	payload := make([]byte, len(data))
	copy(payload, data)
	select {
	case p.queue <- udpOutbound{session: s, payload: payload}:
		p.cfg.AddNetworkFlow(0, uint64(len(data)))
	default:
		// UDP is lossy by design; never block the reader.
		p.cfg.AddConnectRejectedCount(1)
	}
}

// handleResponse routes a datagram coming back from the agent to the client
// endpoint that started the session.
func (p *udpProxy) handleResponse(msg *model.UDPDataMsg) {
	p.mu.Lock()
	s, ok := p.byID[msg.SessionID]
	p.mu.Unlock()
	if !ok {
		return
	}
	s.touch()
	if _, err := p.listener.WriteToUDP(msg.Data, s.clientAddr); err != nil {
		p.h.Errorf("write udp response to %s error: %v", s.clientAddr, err)
		return
	}
	p.cfg.AddNetworkFlow(uint64(len(msg.Data)), 0)
}

func (p *udpProxy) writeLoop() {
	for {
		select {
		case <-p.done:
			return
		case pkt := <-p.queue:
			preferred := pkt.session.getAgentID()
			usedAgent, err := p.sendFn(preferred, pkt.session.id, pkt.payload)
			if err != nil {
				p.h.Errorf("send udp datagram to agent error: %v", err)
				continue
			}
			if usedAgent != "" && usedAgent != preferred {
				pkt.session.setAgentID(usedAgent)
			}
		}
	}
}

func (p *udpProxy) cleanupLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	idle := constants.UDPSessionIdleSeconds * time.Second
	for {
		select {
		case <-p.done:
			return
		case <-ticker.C:
			p.mu.Lock()
			for id, s := range p.byID {
				if s.idleSince() < idle {
					continue
				}
				delete(p.byID, id)
				delete(p.byAddr, s.clientAddr.String())
			}
			p.mu.Unlock()
		}
	}
}

func (p *udpProxy) Close() {
	p.closeOnce.Do(func() {
		close(p.done)
		_ = p.listener.Close()
	})
}
