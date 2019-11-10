package anywhereServer

import (
	"anywhere/conn"
	"anywhere/log"
	"anywhere/model"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

func (s *anyWhereServer) handleNewConnection(c net.Conn) {

	var msg model.RequestMsg
	d := json.NewDecoder(c)

	if err := d.Decode(&msg); err != nil {
		log.GetDefaultLogger().Errorf("unmarshal init pkg error: %v", err)
		_ = c.Close()
	}
	switch msg.ReqType {
	case model.PkgControlConnRegister:
		m, _ := model.ParseControlRegisterPkg(msg.Message)
		agent := NewAgentInfo(m.AgentId, c, make(chan error, 1000))
		if isUpdate := s.RegisterAgent(agent); isUpdate {
			log.GetDefaultLogger().Errorf("rebuild control connection for agent: %v", agent.Id)
		} else {
			log.GetDefaultLogger().Infof("accept control connection from agent: %v", agent.Id)
		}
		go s.handleAdminConnection(agent.Id)
	case model.PkgTunnelBegin:
		m, _ := model.ParseTunnelBeginPkg(msg.Message)
		if !s.isAgentExist(m.AgentId) {
			log.GetDefaultLogger().Errorf("got data conn register pkg from unknown agent %v", m.AgentId)
			_ = c.Close()
		} else {
			log.GetDefaultLogger().Infof("add data conn for %v from agent %v", m.LocalAddr, m.AgentId)
			if err := s.agents[m.AgentId].PutProxyConn(m.LocalAddr, conn.NewBaseConn(c)); err != nil {
				log.GetDefaultLogger().Errorf("put proxy conn to agent error: %v", err)
			}
		}
	default:
		log.GetDefaultLogger().Errorf("unknown msg type %v", msg.ReqType)
		_ = c.Close()

	}

}

func (s *anyWhereServer) handleAdminConnection(id string) {
	if _, ok := s.agents[id]; !ok {
		log.GetDefaultLogger().Errorf("handle on nil admin connection: %v", id)
		return
	}
	msg := &model.RequestMsg{}
	for {
		if err := s.agents[id].AdminConn.Receive(&msg); err != nil {
			log.GetDefaultLogger().Errorf("receive from %v admin conn error: %v, wait client reconnecting", id, err)
			_ = s.agents[id].AdminConn.Close()
			return
		}
		switch msg.ReqType {
		case model.PkgReqHeartBeat:
			m, _ := model.ParseHeartBeatPkg(msg.Message)
			s.SetControlConnHealthy(id, m.SendTime)
		default:
			log.GetDefaultLogger().Errorf("got unknown ReqType: %v from %v", msg.ReqType, id)
			s.CloseControlConnWithResp(id, fmt.Errorf("got unknown ReqType: %v from %v", msg.ReqType, id))
		}
	}
}

func (s *anyWhereServer) SetControlConnHealthy(id string, ackSendTime time.Time) {
	s.agents[id].AdminConn.LastAckRcvTime = time.Now()
	s.agents[id].AdminConn.LastAckSendTime = ackSendTime
	s.agents[id].AdminConn.SetHealthy()
}

func (s *anyWhereServer) CloseControlConnWithResp(id string, err error) {
	m := model.NewResponseMsg(500, err.Error())
	_ = s.agents[id].AdminConn.Send(m)
	_ = s.agents[id].AdminConn.Close()
}
