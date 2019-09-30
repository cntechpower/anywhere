package anywhereServer

import (
	"anywhere/log"
	"anywhere/model"
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
