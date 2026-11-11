package agent

import (
	"net"
	"sync"
	"time"

	"github.com/cntechpower/utils/log"

	"github.com/cntechpower/anywhere/constants"
	"github.com/cntechpower/anywhere/model"
)

// udpManager owns one UDP socket per client session on the agent side. The
// socket is connected to the local service so replies can be read and sent
// back to the server (and finally to the original client).
type udpManager struct {
	mu       sync.Mutex
	sessions map[uint32]*agentUDPSession
	send     func(*model.RequestMsg) error
	closed   bool
}

type agentUDPSession struct {
	id     uint32
	target string
	conn   *net.UDPConn
}

func newUDPManager() *udpManager {
	return &udpManager{
		sessions: make(map[uint32]*agentUDPSession),
	}
}

func (m *udpManager) Handle(msg *model.UDPDataMsg) {
	s := m.getOrCreate(msg)
	if s == nil {
		return
	}
	if _, err := s.conn.Write(msg.Data); err != nil {
		log.Errorf(log.NewHeader("agentUDP"), "write udp data to %s error: %v", s.target, err)
		m.remove(s)
	}
}

func (m *udpManager) getOrCreate(msg *model.UDPDataMsg) *agentUDPSession {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil
	}
	if s, ok := m.sessions[msg.SessionID]; ok {
		if s.target == msg.LocalAddr {
			return s
		}
		// The server restarted and reused the session ID for another target;
		// drop the stale session instead of forwarding to the wrong service.
		delete(m.sessions, msg.SessionID)
		_ = s.conn.Close()
	}
	raddr, err := net.ResolveUDPAddr("udp", msg.LocalAddr)
	if err != nil {
		log.Errorf(log.NewHeader("agentUDP"), "resolve udp addr %s error: %v", msg.LocalAddr, err)
		return nil
	}
	c, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		log.Errorf(log.NewHeader("agentUDP"), "dial udp %s error: %v", msg.LocalAddr, err)
		return nil
	}
	s := &agentUDPSession{id: msg.SessionID, target: msg.LocalAddr, conn: c}
	m.sessions[msg.SessionID] = s
	go m.readLoop(s)
	return s
}

func (m *udpManager) readLoop(s *agentUDPSession) {
	defer m.remove(s)
	buf := make([]byte, constants.UDPMaxDatagramSize)
	idle := time.Duration(constants.UDPSessionIdleSeconds) * time.Second
	for {
		_ = s.conn.SetReadDeadline(time.Now().Add(idle))
		n, err := s.conn.Read(buf)
		if err != nil {
			return
		}
		m.mu.Lock()
		send := m.send
		closed := m.closed
		m.mu.Unlock()
		if closed || send == nil {
			return
		}
		if err := send(model.NewUDPDataRespMsg(s.id, s.target, buf[:n])); err != nil {
			log.Errorf(log.NewHeader("agentUDP"), "send udp response to server error: %v", err)
			return
		}
	}
}

func (m *udpManager) remove(s *agentUDPSession) {
	m.mu.Lock()
	if cur, ok := m.sessions[s.id]; ok && cur == s {
		delete(m.sessions, s.id)
	}
	m.mu.Unlock()
	_ = s.conn.Close()
}

func (m *udpManager) CloseAll() {
	m.mu.Lock()
	m.closed = true
	sessions := make([]*agentUDPSession, 0, len(m.sessions))
	for id, s := range m.sessions {
		sessions = append(sessions, s)
		delete(m.sessions, id)
	}
	m.mu.Unlock()
	for _, s := range sessions {
		_ = s.conn.Close()
	}
}
