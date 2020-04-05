package rpcHandler

import (
	"anywhere/log"
	"anywhere/server/anywhereServer"
	pb "anywhere/server/rpc/definitions"
	"anywhere/util"
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

var (
	ErrServerNotInit = fmt.Errorf("anywhere server not init")
)

type rpcHandlers struct {
	s *anywhereServer.Server
	l *logrus.Entry
}

func GetRpcHandlers(s *anywhereServer.Server) *rpcHandlers {
	return &rpcHandlers{s: s, l: log.GetCustomLogger("grpc_handler")}
}

func (h *rpcHandlers) ListAgent(ctx context.Context, empty *pb.Empty) (*pb.Agents, error) {
	h.l.Infof("calling list agents")
	defer h.l.Infof("called list agents")
	s := anywhereServer.GetServerInstance()
	if s == nil {
		return &pb.Agents{}, ErrServerNotInit
	}
	res := &pb.Agents{
		Agent: make([]*pb.Agent, 0),
	}
	agents := s.ListAgentInfo()
	for _, agent := range agents {
		res.Agent = append(res.Agent, &pb.Agent{
			AgentId:          agent.Id,
			AgentRemoteAddr:  agent.RemoteAddr,
			AgentLastAckRcv:  agent.LastAckRcv,
			AgentLastAckSend: agent.LastAckSend,
		})
	}
	return res, nil
}

func (h *rpcHandlers) AddProxyConfig(ctx context.Context, input *pb.AddProxyConfigInput) (*pb.Empty, error) {
	if input.Config == nil {
		return nil, fmt.Errorf("config not vaild: nil")
	}
	s := anywhereServer.GetServerInstance()
	if s == nil {
		return nil, ErrServerNotInit
	}
	config := input.Config

	if err := util.CheckPortValid(int(config.RemotePort)); err != nil {
		return nil, fmt.Errorf("invalid remoteAddr %v in config, error: %v", config.RemotePort, err)
	}
	if err := util.CheckAddrValid(config.LocalAddr); err != nil {
		return nil, fmt.Errorf("invalid localAddr %v in config, error: %v", config.LocalAddr, err)
	}
	if err := s.AddProxyConfigToAgent(config.AgentId, int(config.RemotePort), config.LocalAddr, config.IsWhiteListOn, config.WhiteCidrList); err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}

func (h *rpcHandlers) ListProxyConfigs(ctx context.Context, input *pb.Empty) (*pb.ListProxyConfigsOutput, error) {
	s := anywhereServer.GetServerInstance()
	if s == nil {
		return nil, ErrServerNotInit
	}
	res := &pb.ListProxyConfigsOutput{
		Config: make([]*pb.ProxyConfig, 0),
	}
	configs := s.ListProxyConfigs()
	for _, config := range configs {
		res.Config = append(res.Config, &pb.ProxyConfig{
			AgentId:       config.AgentId,
			RemotePort:    int64(config.RemotePort),
			LocalAddr:     config.LocalAddr,
			IsWhiteListOn: config.IsWhiteListOn,
			WhiteCidrList: config.WhiteCidrList,
		})
	}
	return res, nil
}

func (h *rpcHandlers) RemoveProxyConfig(ctx context.Context, input *pb.RemoveProxyConfigInput) (*pb.Empty, error) {
	s := anywhereServer.GetServerInstance()
	if s == nil {
		return &pb.Empty{}, ErrServerNotInit
	}
	return &pb.Empty{}, s.RemoveProxyConfigFromAgent(input.AgentId, input.LocalAddr)
}

func (h *rpcHandlers) LoadProxyConfigFile(ctx context.Context, input *pb.Empty) (*pb.Empty, error) {

	s := anywhereServer.GetServerInstance()
	if s == nil {
		return &pb.Empty{}, ErrServerNotInit
	}
	return &pb.Empty{}, s.LoadProxyConfigFile()
}

func (h *rpcHandlers) SaveProxyConfigToFile(ctx context.Context, input *pb.Empty) (*pb.Empty, error) {
	s := anywhereServer.GetServerInstance()
	if s == nil {
		return &pb.Empty{}, ErrServerNotInit
	}
	return &pb.Empty{}, s.SaveConfigToFile()
}

func (h *rpcHandlers) ListConns(ctx context.Context, input *pb.ListConnsInput) (*pb.Conns, error) {
	h.l.Infof("calling list conns")
	defer h.l.Infof("called list conns")
	agentConnsMap, err := h.s.ListJoinedConns(input.AgentId)
	if err != nil {
		return nil, err
	}
	res := &pb.Conns{
		Conn: make([]*pb.Conn, 0),
	}

	for agentId, agentConns := range agentConnsMap {
		for _, conn := range agentConns {
			res.Conn = append(res.Conn, &pb.Conn{
				AgentId:       agentId,
				ConnId:        int64(conn.ConnId),
				SrcRemoteAddr: conn.SrcRemoteAddr,
				SrcLocalAddr:  conn.SrcLocalAddr,
				DstRemoteAddr: conn.DstRemoteAddr,
				DstLocalAddr:  conn.DstLocalAddr,
			})
		}

	}

	return res, nil
}

func (h *rpcHandlers) KillConnById(ctx context.Context, input *pb.KillConnByIdInput) (*pb.Empty, error) {
	h.l.Infof("calling kill conn %v on agent %v", input.ConnId, input.AgentId)
	defer h.l.Infof("called kill conn %v on agent %v", input.ConnId, input.AgentId)
	return &pb.Empty{}, h.s.KillJoinedConnById(input.AgentId, int(input.ConnId))
}

func (h *rpcHandlers) KillAllConns(ctx context.Context, empty *pb.Empty) (*pb.Empty, error) {
	h.l.Infof("calling flush conns")
	defer h.l.Infof("called flush conns")
	h.s.FlushJoinedConns()
	return &pb.Empty{}, nil
}
