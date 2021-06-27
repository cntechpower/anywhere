package server

import (
	"fmt"
	"net"

	"github.com/cntechpower/anywhere/dao/config"
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

func (s *Server) AddProxyConfig(userName, zoneName string, remotePort int, localAddr string, isWhiteListOn bool, whiteList string) error {
	pkg, err := model.NewProxyConfig(userName, zoneName, remotePort, localAddr, isWhiteListOn, whiteList)
	if err != nil {
		return err
	}
	if err := s.AddProxyConfigByModel(pkg); err != nil {
		return err
	}
	return config.Add(pkg)

}

func (s *Server) AddProxyConfigByModel(config *model.ProxyConfig) error {
	if !s.isZoneExist(config.UserName, config.ZoneName) {
		return fmt.Errorf("group %v not exist", config.ZoneName)
	}
	s.agentsRwMutex.Lock()
	defer s.agentsRwMutex.Unlock()
	if err := s.zones[config.UserName][config.ZoneName].AddProxyConfig(config); err != nil {
		return err
	}
	return nil
}

func (s *Server) RemoveProxyConfig(userName string, zoneName string, remotePort int, localAddr string) (err error) {
	if !s.isZoneExist(userName, zoneName) {
		return fmt.Errorf("zoneName %v not exist", zoneName)
	}
	if err := util.CheckAddrValid(localAddr); err != nil {
		return fmt.Errorf("invalid localAddr %v, error: %v", localAddr, err)
	}
	s.agentsRwMutex.Lock()
	defer s.agentsRwMutex.Unlock()
	if err = s.zones[userName][zoneName].RemoveProxyConfig(remotePort, localAddr); err != nil {
		return err
	}
	return

}

func (s *Server) RemoveProxyConfigById(id uint) (err error) {
	pc, err := config.GetById(int64(id))
	if err != nil {
		return
	}
	err = s.RemoveProxyConfig(pc.UserName, pc.ZoneName, pc.RemotePort, pc.LocalAddr)
	if err != nil {
		return
	}
	err = config.Remove(id)
	return
}
