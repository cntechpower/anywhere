package server

import (
	"fmt"

	"github.com/cntechpower/anywhere/dao/config"
	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/util"
)

func (s *Server) AddProxyConfig(userName, zoneName string, remotePort int, localAddr string, isWhiteListOn bool, whiteList, listenType string) error {
	cfg, err := model.NewProxyConfig(userName, zoneName, remotePort, localAddr, isWhiteListOn, whiteList, listenType)
	if err != nil {
		return err
	}
	// Persist first so the row ID is known before the config is published to
	// readers; the embedded config must not be mutated after publication.
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("save proxy config: %w", err)
	}
	if err := s.AddProxyConfigByModel(cfg); err != nil {
		_ = config.Remove(cfg.ID)
		return err
	}
	return nil
}

func (s *Server) AddProxyConfigByModel(cfg *model.ProxyConfig) error {
	z, ok := s.getZone(cfg.UserName, cfg.ZoneName)
	if !ok {
		return fmt.Errorf("zone %s of user %s does not exist", cfg.ZoneName, cfg.UserName)
	}
	return z.AddProxyConfig(cfg)
}

func (s *Server) RemoveProxyConfig(userName, zoneName string, remotePort int, localAddr string) error {
	z, ok := s.getZone(userName, zoneName)
	if !ok {
		return fmt.Errorf("zone %s of user %s does not exist", zoneName, userName)
	}
	if err := util.CheckAddrValid(localAddr); err != nil {
		return fmt.Errorf("invalid localAddr %s: %w", localAddr, err)
	}
	if err := z.RemoveProxyConfig(remotePort, localAddr); err != nil {
		return err
	}
	return config.RemoveByKey(userName, zoneName, remotePort, localAddr)
}

func (s *Server) RemoveProxyConfigById(id uint) error {
	pc, err := config.GetById(int64(id))
	if err != nil {
		return err
	}
	if err := s.RemoveProxyConfig(pc.UserName, pc.ZoneName, pc.RemotePort, pc.LocalAddr); err != nil {
		return err
	}
	return config.Remove(uint(id))
}
