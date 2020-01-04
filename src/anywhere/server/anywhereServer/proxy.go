package anywhereServer

import (
	"anywhere/log"
	"anywhere/model"
	"anywhere/util"
	"fmt"
	"net"
)

func (s *anyWhereServer) listenPort(addr string) *net.TCPListener {
	rAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.GetDefaultLogger().Errorf("parse proxy port error: %v", err)
	}
	ln, err := net.ListenTCP("tcp", rAddr)
	if err != nil {
		log.GetDefaultLogger().Errorf("listen proxy port error: %v", err)
	}
	return ln
}

func (s *anyWhereServer) AddProxyConfigToAgent(agentId string, remotePort int, localAddr string, isWhiteListOn bool, whiteList string) error {
	if !s.isAgentExist(agentId) {
		return fmt.Errorf("agent %v not exist", agentId)
	}
	pkg, err := model.NewProxyConfig(remotePort, localAddr, isWhiteListOn, whiteList)
	if err != nil {
		return err
	}
	s.agents[agentId].AddProxyConfig(pkg)
	return nil
}

func (s *anyWhereServer) RemoveProxyConfigFromAgent(agentId, localAddr string) error {
	if !s.isAgentExist(agentId) {
		return fmt.Errorf("agent %v not exist", agentId)
	}
	if err := util.CheckAddrValid(localAddr); err != nil {
		return fmt.Errorf("invalid localAddr %v, error: %v", localAddr, err)
	}
	return s.agents[agentId].RemoveProxyConfig(localAddr)

}
