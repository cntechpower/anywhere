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
		agent := NewAgentInfo(m.AgentId, s.serverId, c)
		if isUpdate := s.RegisterAgent(agent); isUpdate {
			log.Info("rebuild control connection for agent: %v", agent.Id)
		}
		go s.handleAdminConnection(agent.Id)
	case model.PkgDataConnRegister:
		m, _ := model.ParseDataConnRegisterPkg(msg.Message)
		if !s.isAgentExist(m.AgentId) {
			log.Error("got data conn register pkg from unknown agent %v", m.AgentId)
			_ = c.Close()
		} else {
			log.Info("add data conn for agent %v", m.AgentId)
			s.addDataConnToAgent(m.AgentId, c)
		}
	default:
		log.Error("agent %v not register", msg.From)
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
			s.agents[id].AdminConn.Close()
			return
		}
		switch msg.ReqType {
		case model.PkgReqNewproxy:
			m, _ := model.ParseProxyConfig(msg.Message)
			log.Info("got PkgReqNewproxy: %v, %v", m.RemoteAddr, m.LocalAddr)
			s.addProxyConfig(id, m.RemoteAddr, m.LocalAddr)
		case model.PkgReqHeartBeat:
			m, _ := model.ParseHeartBeatPkg(msg.Message)
			log.Info("got PkgReqHeartBeat from %v, sendTime: %v", m.RemoteAddr, m.SendTime.String())
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
	s.agents[id].AdminConn.Close()
}

func (s *anyWhereServer) addDataConnToAgent(id string, c net.Conn) {
	s.agents[id].DataConn = append(s.agents[id].DataConn,
		DataConnStatus{
			BaseConn: conn.NewBaseConn(c),
			InUsed:   false,
		},
	)
}

func (s *anyWhereServer) getDataConnToAgent(id string) (*conn.BaseConn, error) {
	if !s.isAgentExist(id) {
		return nil, fmt.Errorf("not init")
	}
	for _, c := range s.agents[id].DataConn {
		if !c.InUsed {
			return c.BaseConn, nil
		}
	}
	return nil, fmt.Errorf("no data conn available")
}
