package anywhereServer

import (
	"anywhere/conn"
	"anywhere/log"
	"anywhere/model"
	"encoding/json"
	"net"
)

func (s *Server) handleNewConnection(c net.Conn) {
	h := log.NewHeader("handleNewAgentConn")
	var msg model.RequestMsg
	d := json.NewDecoder(c)

	if err := d.Decode(&msg); err != nil {
		log.Errorf(h, "unmarshal init pkg from %s error: %v", c.RemoteAddr(), err)
		_ = c.Close()
		return
	}
	switch msg.ReqType {
	case model.PkgControlConnRegister:
		m, _ := model.ParseControlRegisterPkg(msg.Message)
		agent := NewAgentInfo(m.AgentId, c, make(chan error, 1000))
		if isUpdate := s.RegisterAgent(agent); isUpdate {
			log.Errorf(h, "rebuild control connection for agent: %v", agent.Id)
		} else {
			log.Infof(h, "accept control connection from agent: %v", agent.Id)
		}
		log.Infof(h, "got conn from : %v", c.RemoteAddr())
		//go s.handleAdminConnection(agent.Id)
		go agent.handleAdminConnection()
	case model.PkgTunnelBegin:
		m, _ := model.ParseTunnelBeginPkg(msg.Message)
		if !s.isAgentExist(m.AgentId) {
			log.Errorf(h, "got data conn register pkg from unknown agent %v", m.AgentId)
			_ = c.Close()
		} else {
			log.Infof(h, "add data conn for %v from agent %v", m.LocalAddr, m.AgentId)
			if err := s.agents[m.AgentId].PutProxyConn(m.LocalAddr, conn.NewBaseConn(c)); err != nil {
				log.Errorf(h, "put proxy conn to agent error: %v", err)
			}
		}
	default:
		log.Errorf(h, "unknown msg type %v from %v", msg.ReqType, c.RemoteAddr())
		_ = c.Close()

	}

}
