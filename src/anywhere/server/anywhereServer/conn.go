package anywhereServer

import (
	"anywhere/log"
	"anywhere/model"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

func (s *anyWhereServer) handleNewConnection(c net.Conn, funcOnError func(c net.Conn, err error)) {

	var msg model.RequestMsg
	d := json.NewDecoder(c)

	if err := d.Decode(&msg); err != nil {
		funcOnError(c, fmt.Errorf("unmarshal init pkg error: %v", err))
	}
	switch msg.ReqType {
	case model.PkgRegister:
		m, _ := model.ParseRegisterPkg(msg.Message)
		agent := NewAgentInfo(m.AgentId, s.serverId, c)
		if isUpdate := s.RegisterAgent(agent); isUpdate {
			log.Info("rebuild control connection for agent: %v", agent.Id)
		}
		go s.handleAdminConnection(agent.Id)
	default:
		funcOnError(c, fmt.Errorf("agent %v not register", msg.From))

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
		case model.PkgReqHeartBeat:
			m, _ := model.ParseHeartBeatPkg(msg.Message)
			log.Info("got PkgReqHeartBeat from %v, sendTime: %v", m.RemoteAddr, m.SendTime.String())
			s.SetControlConnHealthy(id, m.SendTime)
		default:
			log.Error("got unknown ReqType: %v from %v", msg.ReqType, id)
			s.CloseControlConnWithResp(id, fmt.Errorf("got unknown ReqType: %v from %v", msg.ReqType, id))
		}
		rsp := model.NewResponseMsg(200, "got it")
		if err := s.agents[id].AdminConn.Send(rsp); err != nil {
			log.Error("send response to %v error: %v", id, err)
			s.CloseControlConnWithResp(id, err)
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
