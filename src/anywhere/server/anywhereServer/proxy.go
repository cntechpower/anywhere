package anywhereServer

import (
	"anywhere/log"
	"anywhere/model"
	"anywhere/util"
	"fmt"
	"net"
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

func (s *Server) AddProxyConfigToAgent(agentId string, remotePort int, localAddr string, isWhiteListOn bool, whiteList string) error {

	pkg, err := model.NewProxyConfig(agentId, remotePort, localAddr, isWhiteListOn, whiteList)
	if err != nil {
		return err
	}
	return s.AddProxyConfigToAgentByModel(pkg)

}

func (s *Server) AddProxyConfigToAgentByModel(config *model.ProxyConfig) error {
	if !s.isAgentExist(config.AgentId) {
		return fmt.Errorf("agent %v not exist", config.AgentId)
	}
	return s.agents[config.AgentId].AddProxyConfig(config)
}

func (s *Server) RemoveProxyConfigFromAgent(agentId, localAddr string) error {
	if !s.isAgentExist(agentId) {
		return fmt.Errorf("agent %v not exist", agentId)
	}
	if err := util.CheckAddrValid(localAddr); err != nil {
		return fmt.Errorf("invalid localAddr %v, error: %v", localAddr, err)
	}
	return s.agents[agentId].RemoveProxyConfig(localAddr)

}
