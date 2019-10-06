package anywhereServer

import (
	"anywhere/log"
	"anywhere/model"
	"anywhere/util"
	_tls "crypto/tls"
	"net"
	"strconv"
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
	rAddr, _ := net.ResolveTCPAddr("tcp", remoteAddr)
	s.listenPort(strconv.Itoa(rAddr.Port))
}

func (s *anyWhereServer) listenPort(port string) net.Listener {
	addr, err := util.GetAddrByIpPort("0.0.0.0", port)
	if err != nil {
		log.Error("listen proxy port error: %v", err)
	}
	ln, err := _tls.Listen("tcp", addr.String(), s.credential)
	if err != nil {
		log.Error("listen proxy port error: %v", err)
	}
	return ln
}
