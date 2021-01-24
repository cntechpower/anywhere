package server

import (
	"fmt"
	"net"

	"github.com/cntechpower/anywhere/server/conf"

	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/util"
	"github.com/cntechpower/utils/log"
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

func (s *Server) AddProxyConfig(userName, groupName string, remotePort int, localAddr string, isWhiteListOn bool, whiteList string) error {
	pkg, err := model.NewProxyConfig(userName, groupName, remotePort, localAddr, isWhiteListOn, whiteList)
	if err != nil {
		return err
	}
	if err := s.AddProxyConfigByModel(pkg); err != nil {
		return err
	}
	return conf.Add(pkg)

}

func (s *Server) AddProxyConfigByModel(config *model.ProxyConfig) error {
	if !s.isGroupExist(config.UserName, config.GroupName) {
		return fmt.Errorf("group %v not exist", config.GroupName)
	}
	s.agentsRwMutex.Lock()
	defer s.agentsRwMutex.Unlock()
	if err := s.groups[config.UserName][config.GroupName].AddProxyConfig(config); err != nil {
		return err
	}
	return nil
}

func (s *Server) RemoveProxyConfig(userName string, group string, remotePort int, localAddr string) error {
	if !s.isGroupExist(userName, group) {
		return fmt.Errorf("group %v not exist", group)
	}
	if err := util.CheckAddrValid(localAddr); err != nil {
		return fmt.Errorf("invalid localAddr %v, error: %v", localAddr, err)
	}
	s.agentsRwMutex.Lock()
	defer s.agentsRwMutex.Unlock()
	if err := s.groups[userName][group].RemoveProxyConfig(remotePort, localAddr); err != nil {
		return err
	}
	return conf.Remove(userName, group, remotePort)

}
