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
	for _, c := range s.agents[id].DataConn {
		if !c.InUsed {
			c.InUsed = true
			return c.BaseConn, nil
		}
	}
	return nil, fmt.Errorf("no data conn available")
}

func (s *anyWhereServer) releaseDataConn(id string, bc *conn.BaseConn) {
	log.Info("release %v data conn", bc.RemoteAddr())
	for _, c := range s.agents[id].DataConn {
		if c.BaseConn == bc {
			c.InUsed = false
			return
		}
	}
	log.Info("no conn to release")
}

func (s *anyWhereServer) handelTunnelConnection(ln *net.TCPListener, localAddr, agentId string) {
	for {
		c, err := ln.AcceptTCP()
		if err != nil {
			log.Error("got conn from %v err: %v", ln.Addr(), err)
			continue
		}
		//go s.tunnelHandlerWithoutPool(c, localAddr, agentId)
		go s.tunnelHandlerWithPool(c, localAddr, agentId)

	}
}

func (s *anyWhereServer) tunnelHandlerWithoutPool(c net.Conn, localAddr, agentId string) {
	dst, err := s.getAvailDataConn(agentId)
	if err != nil {
		log.Error("get conn error %v", err)
		_ = c.Close()
		return
	}

	//call agent to join conn
	p := model.NewTunnelBeginMsg(localAddr)
	pkg := model.NewRequestMsg("0.0.1", model.PkgDataConnTunnel, s.serverId, "", p)
	pByte, _ := json.Marshal(pkg)
	_, err = dst.Write(pByte)
	if err != nil {
		log.Error("send tunnel begin to client error: %v", err)
		return
	}

	//server join conn
	//dst.CancelHeartBeatSend()
	//dst.CancelHeartBeatReceive()
	//dst.StopRcvChan <- struct{}{}
	conn.JoinConn(dst.Conn, c)
	s.releaseDataConn(agentId, dst)

}

func (s *anyWhereServer) tunnelHandlerWithPool(c net.Conn, localAddr, agentId string) {
	dst, err := conn.GetFromPool(agentId)
	if err != nil {
		log.Error("get conn error %v", err)
		_ = c.Close()
		return
	}

	//call agent to join conn
	p := model.NewTunnelBeginMsg(localAddr)
	pkg := model.NewRequestMsg("0.0.1", model.PkgDataConnTunnel, s.serverId, "", p)
	pByte, _ := json.Marshal(pkg)
	_, err = dst.Write(pByte)
	if err != nil {
		log.Error("send tunnel begin to client error: %v", err)
		return
	}

	//clear conn rcv buffer
	//_, _ = ioutil.ReadAll(dst)

	//server join conn
	//dst.CancelHeartBeatSend()
	//dst.CancelHeartBeatReceive()
	//dst.StopRcvChan <- struct{}{}
	conn.JoinConn(dst.Conn, c)

}
