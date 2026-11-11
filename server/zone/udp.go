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

type udpOutbound struct {
	sessionID uint32
	payload   []byte
}

type udpSession struct {
	id         uint32
	clientAddr *net.UDPAddr
	lastActive atomic.Int64 // unix nano
}

func (s *udpSession) touch() { s.lastActive.Store(time.Now().UnixNano()) }

func (s *udpSession) idleSince() time.Duration {
	return time.Since(time.Unix(0, s.lastActive.Load()))
}

// udpProxy keeps the state of one UDP listener: which client endpoint maps to
// which session ID and how to send datagrams towards the agent.
type udpProxy struct {
	cfg      *ProxyConfigStats
	listener *net.UDPConn
	sendFn   func(sessionID uint32, data []byte) error
	h        *log.Header

	mu     sync.Mutex
	byID   map[uint32]*udpSession
	byAddr map[string]*udpSession
	nextID atomic.Uint32

	queue     chan udpOutbound
	done      chan struct{}
	closeOnce sync.Once
}

func newUDPProxy(cfg *ProxyConfigStats, listener *net.UDPConn, sendFn func(sessionID uint32, data []byte) error) *udpProxy {
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
	id := p.nextID.Add(1)
	s := &udpSession{id: id, clientAddr: client}
	s.touch()
	p.byID[id] = s
	p.byAddr[key] = s
	p.cfg.AddConnectCount(1)
	return s
}

func (p *udpProxy) handleDatagram(data []byte, client *net.UDPAddr) {
	s := p.session(client)
	payload := make([]byte, len(data))
	copy(payload, data)
	select {
	case p.queue <- udpOutbound{sessionID: s.id, payload: payload}:
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
	}
}

func (p *udpProxy) writeLoop() {
	for {
		select {
		case <-p.done:
			return
		case pkt := <-p.queue:
			if err := p.sendFn(pkt.sessionID, pkt.payload); err != nil {
				p.h.Errorf("send udp datagram to agent error: %v", err)
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
