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
		log.Error("unmarshal init pkg error: %v", err)
		_ = c.Close()
	}
	switch msg.ReqType {
	case model.PkgControlConnRegister:
		m, _ := model.ParseControlRegisterPkg(msg.Message)
		agent := NewAgentInfo(m.AgentId, c, make(chan error, 1000))
		if isUpdate := s.RegisterAgent(agent); isUpdate {
			log.Info("rebuild control connection for agent: %v", agent.Id)
		} else {
			log.Info("accept control connection from agent: %v", agent.Id)
		}
		go s.handleAdminConnection(agent.Id)
	case model.PkgTunnelBegin:
		m, _ := model.ParseTunnelBeginPkg(msg.Message)
		if !s.isAgentExist(m.AgentId) {
			log.Error("got data conn register pkg from unknown agent %v", m.AgentId)
			_ = c.Close()
		} else {
			log.Info("add data conn for %v from agent %v", m.LocalAddr, m.AgentId)
			if err := s.agents[m.AgentId].PutProxyConn(m.LocalAddr, conn.NewBaseConn(c)); err != nil {
				log.Error("put proxy conn to agent error: %v", err)
			}
		}
	default:
		log.Error("unknown msg type %v", msg.ReqType)
		_ = c.Close()

	}

}

func (s *anyWhereServer) handleAdminConnection(id string) {
	if _, ok := s.agents[id]; !ok {
		log.Fatal("handle on nil admin connection: %v", id)
	}
	msg := &model.RequestMsg{}
	for {
		if err := s.agents[id].AdminConn.Receive(&msg); err != nil {
			log.Error("receive from %v admin conn error: %v, wait client reconnecting", id, err)
			_ = s.agents[id].AdminConn.Close()
			return
		}
		switch msg.ReqType {
		case model.PkgReqHeartBeat:
			m, _ := model.ParseHeartBeatPkg(msg.Message)
			s.SetControlConnHealthy(id, m.SendTime)
		default:
			log.Error("got unknown ReqType: %v from %v", msg.ReqType, id)
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
