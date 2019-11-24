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
		log.GetDefaultLogger().Errorf("parse proxy port error: %v", err)
	}
	ln, err := net.ListenTCP("tcp", rAddr)
	if err != nil {
		log.GetDefaultLogger().Errorf("listen proxy port error: %v", err)
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
	return nil
}

func (s *anyWhereServer) RemoveProxyConfigFromAgent(agentId string, localIp, localPort string) error {
	if !s.isAgentExist(agentId) {
		return fmt.Errorf("agent %v not exist", agentId)
	}
	return s.agents[agentId].RemoveProxyConfig(fmt.Sprintf("%v:%v", localIp, localPort))

}
