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
			//s.handleDataConnection(m.AgentId,c)
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
			_ = s.agents[id].AdminConn.Close()
			return
		}
		switch msg.ReqType {
		case model.PkgReqNewproxy:
			m, _ := model.ParseProxyConfig(msg.Message)
			log.Info("got PkgReqNewproxy: %v, %v", m.RemoteAddr, m.LocalAddr)
			s.addProxyConfig(id, m.RemoteAddr, m.LocalAddr)
			go s.handelTunnelConnection(s.listenPort(m.RemoteAddr), m.LocalAddr, id)
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
	s.agents[id].AdminConn.Close()
}

func (s *anyWhereServer) addDataConnToAgent(id string, c net.Conn) {
	baseConn := conn.NewBaseConn(c)
	s.agents[id].DataConn = append(s.agents[id].DataConn,
		&DataConn{
			BaseConn: baseConn,
			InUsed:   false,
		},
	)
	conn.HeartBeatRcvLoop(baseConn, s.destroyDataConn)
}

//
//func HeartBeatRcvLoop(c *conn.BaseConn, funcOnFail func(c *conn.BaseConn)) {
//	msg := &model.RequestMsg{}
//	go func(c *conn.BaseConn) {
//		for {
//			select {
//			case <-c.StopRcvChan:
//				return
//			default:
//			}
//			if err := c.Receive(&msg); err != nil {
//				log.Error("receive from data conn %v  error: %v, close this data conn", c.RemoteAddr(), err)
//				_ = c.Close()
//				funcOnFail(c)
//				return
//			}
//			switch msg.ReqType {
//			case model.PkgReqHeartBeat:
//				m, _ := model.ParseHeartBeatPkg(msg.Message)
//				c.LastAckSendTime = m.SendTime
//				c.LastAckRcvTime = time.Now()
//				c.SetHealthy()
//			case model.PkgDataConnTunnel:
//				log.Info("got data conn tunnel, exit handleDataConnection for %v", c.RemoteAddr())
//				return
//			default:
//				log.Error("got unknown ReqType: %v from %v", msg.ReqType, c.RemoteAddr())
//				_ = c.Close()
//				funcOnFail(c)
//				return
//			}
//		}
//	}(c)
//}

func (s *anyWhereServer) destroyDataConn(c *conn.BaseConn) {
	for _, agentDataConn := range s.agents {
		for idx, c1 := range agentDataConn.DataConn {
			if c == c1.BaseConn {
				if idx+1 == len(agentDataConn.DataConn) {
					agentDataConn.DataConn = make([]*DataConn, 0)
					return
				}
				agentDataConn.DataConn = append(agentDataConn.DataConn[:idx], agentDataConn.DataConn[idx+1:]...)
			}
		}
	}
}
