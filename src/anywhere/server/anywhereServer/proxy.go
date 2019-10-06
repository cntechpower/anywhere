package anywhereServer

import (
	"anywhere/conn"
	"anywhere/log"
	"anywhere/model"
	"encoding/json"
	"fmt"
	"net"
)

func (s *anyWhereServer) addProxyConfig(id, remoteAddr, localAddr string) {
	if _, ok := s.agents[id]; !ok {
		log.Error("add proxy to a not exist agent %v", id)
		return
	}
	s.agents[id].ProxyConfigs = append(s.agents[id].ProxyConfigs, model.ProxyConfig{
		RemoteAddr: remoteAddr,
		LocalAddr:  localAddr,
	})
}

func (s *anyWhereServer) listenPort(addr string) *net.TCPListener {
	rAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Error("parse proxy port error: %v", err)
	}
	ln, err := net.ListenTCP("tcp", rAddr)
	if err != nil {
		log.Error("listen proxy port error: %v", err)
	}
	return ln
}

func (s *anyWhereServer) getAvailDataConn(id string) (*conn.BaseConn, error) {
	for idx, c := range s.agents[id].DataConn {
		if !c.InUsed {
			//TODO: set InUsed not work, so remove this conn from dataConn
			s.agents[id].DataConn = append(s.agents[id].DataConn[:idx], s.agents[id].DataConn[idx+1:]...)
			//c.InUsed = true
			//c.BaseConn.Send()
			return c.BaseConn, nil
		}
	}
	return nil, fmt.Errorf("no data conn available")
}

func (s *anyWhereServer) handelProxyConn(ln *net.TCPListener, localAddr, agentId string) {
	for {
		c, err := ln.AcceptTCP()
		if err != nil {
			log.Error("got conn from %v err: %v", ln.Addr(), err)
			continue
		}
		dst, err := s.getAvailDataConn(agentId)
		if err != nil {
			log.Error("get conn error %v", err)
			_ = c.Close()
			return
		}
		log.Info("got proxy conn from %v to %v", c.RemoteAddr(), ln.Addr())
		p := model.NewTunnelBeginMsg(localAddr)
		pkg := model.NewRequestMsg("0.0.1", model.PkgDataConnTunnel, s.serverId, "", p)
		pByte, _ := json.Marshal(pkg)
		_, err = dst.Write(pByte)
		if err != nil {
			log.Error("send tunnel begin to client error: %v", err)
		} else {
			go conn.JoinConn(c, dst.Conn)
		}

	}
}
