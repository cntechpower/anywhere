package anywhereServer

import (
	"anywhere/log"
	"anywhere/model"
	"fmt"
	"net"
)

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

func (s *anyWhereServer) AddProxyConfigToAgent(agentId string, remotePort int, localIp string, localPort int) error {
	if !s.isAgentExist(agentId) {
		return fmt.Errorf("agent %v not exist", agentId)
	}
	pkg, err := model.NewProxyConfigMsg(remotePort, localIp, localPort)
	if err != nil {
		return err
	}
	s.agents[agentId].AddProxyConfig(pkg)
	go s.agents[agentId].ProxyConfigHandleLoop()
	return nil
}
