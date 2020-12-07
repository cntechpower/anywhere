package server

import (
	"fmt"
	"net"

	"github.com/cntechpower/anywhere/server/conf"

	"github.com/cntechpower/anywhere/log"
	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/util"
)

func (s *Server) listenPort(addr string) *net.TCPListener {
	h := log.NewHeader("serverListenPort")
	rAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Errorf(h, "parse proxy port error: %v", err)
	}
	ln, err := net.ListenTCP("tcp", rAddr)
	if err != nil {
		log.Errorf(h, "listen proxy port error: %v", err)
	}
	return ln
}

func (s *Server) AddProxyConfigToAgent(userName, agentId string, remotePort int, localAddr string, isWhiteListOn bool, whiteList string) error {

	pkg, err := model.NewProxyConfig(userName, agentId, remotePort, localAddr, isWhiteListOn, whiteList)
	if err != nil {
		return err
	}
	if err := s.AddProxyConfigToAgentByModel(pkg); err != nil {
		return err
	}
	return conf.Add(pkg)

}

func (s *Server) AddProxyConfigToAgentByModel(config *model.ProxyConfig) error {
	if !s.isAgentExist(config.UserName, config.AgentId) {
		return fmt.Errorf("agent %v not exist", config.AgentId)
	}
	s.agentsRwMutex.Lock()
	defer s.agentsRwMutex.Unlock()
	if err := s.agents[config.UserName][config.AgentId].AddProxyConfig(config); err != nil {
		return err
	}
	return nil
}

func (s *Server) RemoveProxyConfigFromAgent(userName string, remotePort int, agentId, localAddr string) error {
	if !s.isAgentExist(userName, agentId) {
		return fmt.Errorf("agent %v not exist", agentId)
	}
	if err := util.CheckAddrValid(localAddr); err != nil {
		return fmt.Errorf("invalid localAddr %v, error: %v", localAddr, err)
	}
	s.agentsRwMutex.Lock()
	defer s.agentsRwMutex.Unlock()
	if err := s.agents[userName][agentId].RemoveProxyConfig(remotePort, localAddr); err != nil {
		return err
	}
	return conf.Remove(userName, agentId, remotePort)

}
