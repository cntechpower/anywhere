package handler

import (
	"context"
	"fmt"

	"github.com/cntechpower/anywhere/log"
	pb "github.com/cntechpower/anywhere/server/rpc/definitions"
	"github.com/cntechpower/anywhere/server/server"
	"github.com/cntechpower/anywhere/util"
)

var (
	ErrServerNotInit = fmt.Errorf("anywhere server not init")
)

type rpcHandlers struct {
	s         *server.Server
	logHeader *log.Header
}

func GetRpcHandlers(s *server.Server) *rpcHandlers {
	return &rpcHandlers{
		s:         s,
		logHeader: log.NewHeader("rpcHandler"),
	}
}

func (h *rpcHandlers) ListAgent(ctx context.Context, empty *pb.Empty) (*pb.Agents, error) {
	s := server.GetServerInstance()
	if s == nil {
		return &pb.Agents{}, ErrServerNotInit
	}
	res := &pb.Agents{
		Agent: make([]*pb.Agent, 0),
	}
	agents := s.ListAgentInfo()
	for _, agent := range agents {
		res.Agent = append(res.Agent, &pb.Agent{
			AgentUserName:         agent.UserName,
			AgentId:               agent.Id,
			AgentRemoteAddr:       agent.RemoteAddr,
			AgentLastAckRcv:       agent.LastAckRcv,
			AgentLastAckSend:      agent.LastAckSend,
			AgentProxyConfigCount: int64(agent.ProxyConfigCount),
		})
	}
	return res, nil
}

func (h *rpcHandlers) AddProxyConfig(ctx context.Context, input *pb.AddProxyConfigInput) (*pb.Empty, error) {
	if input.Config == nil {
		return nil, fmt.Errorf("config not vaild: nil")
	}
	s := server.GetServerInstance()
	if s == nil {
		return &pb.Empty{}, ErrServerNotInit
	}
	config := input.Config

	if err := util.CheckPortValid(int(config.RemotePort)); err != nil {
		return &pb.Empty{}, fmt.Errorf("invalid remoteAddr %v in config, error: %v", config.RemotePort, err)
	}
	if err := util.CheckAddrValid(config.LocalAddr); err != nil {
		return &pb.Empty{}, fmt.Errorf("invalid localAddr %v in config, error: %v", config.LocalAddr, err)
	}
	if err := s.AddProxyConfigToAgent(config.Username, config.AgentId, int(config.RemotePort), config.LocalAddr, config.IsWhiteListOn, config.WhiteCidrList); err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}

func (h *rpcHandlers) ListProxyConfigs(ctx context.Context, input *pb.Empty) (*pb.ListProxyConfigsOutput, error) {
	s := server.GetServerInstance()
	if s == nil {
		return nil, ErrServerNotInit
	}
	res := &pb.ListProxyConfigsOutput{
		Config: make([]*pb.ProxyConfig, 0),
	}
	configs := s.ListProxyConfigs()
	for _, config := range configs {
		res.Config = append(res.Config, &pb.ProxyConfig{
			AgentId:                         config.AgentId,
			RemotePort:                      int64(config.RemotePort),
			LocalAddr:                       config.LocalAddr,
			IsWhiteListOn:                   config.IsWhiteListOn,
			WhiteCidrList:                   config.WhiteCidrList,
			NetworkFlowLocalToRemoteInBytes: int64(config.NetworkFlowLocalToRemoteInBytes),
			NetworkFlowRemoteToLocalInBytes: int64(config.NetworkFlowRemoteToLocalInBytes),
		})
	}
	return res, nil
}

func (h *rpcHandlers) RemoveProxyConfig(ctx context.Context, input *pb.RemoveProxyConfigInput) (*pb.Empty, error) {
	s := server.GetServerInstance()
	if s == nil {
		return &pb.Empty{}, ErrServerNotInit
	}
	return &pb.Empty{}, s.RemoveProxyConfigFromAgent(input.UserName, int(input.RemotePort), input.AgentId, input.LocalAddr)
}

func (h *rpcHandlers) LoadProxyConfigFile(ctx context.Context, input *pb.Empty) (*pb.Empty, error) {

	s := server.GetServerInstance()
	if s == nil {
		return &pb.Empty{}, ErrServerNotInit
	}
	return &pb.Empty{}, s.LoadProxyConfigFile()
}

func (h *rpcHandlers) SaveProxyConfigToFile(ctx context.Context, input *pb.Empty) (*pb.Empty, error) {
	s := server.GetServerInstance()
	if s == nil {
		return &pb.Empty{}, ErrServerNotInit
	}
	return &pb.Empty{}, nil //TODO: use config auto save and remove this api.
}

func (h *rpcHandlers) ListConns(ctx context.Context, input *pb.ListConnsInput) (*pb.Conns, error) {
	agentConnsMap, err := h.s.ListJoinedConns("", input.AgentId)
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
	return &pb.Empty{}, h.s.KillJoinedConnById(input.UserName, input.AgentId, int(input.ConnId))
}

func (h *rpcHandlers) KillAllConns(ctx context.Context, empty *pb.Empty) (*pb.Empty, error) {
	h.s.FlushJoinedConns()
	return &pb.Empty{}, nil
}

func (h *rpcHandlers) UpdateProxyConfigWhiteList(ctx context.Context, input *pb.UpdateProxyConfigWhiteListInput) (*pb.Empty, error) {
	return &pb.Empty{}, h.s.UpdateProxyConfigWhiteList(input.UserName, int(input.RemotePort), input.AgentId, input.LocalAddr, input.WhiteCidrs, input.WhiteListEnable)
}

func (h *rpcHandlers) GetSummary(ctx context.Context, empty *pb.Empty) (*pb.GetSummaryOutput, error) {
	s := h.s.GetSummary()
	res := &pb.GetSummaryOutput{
		AgentCount:                  int64(s.AgentTotalCount),
		ProxyCount:                  int64(s.ProxyConfigTotalCount),
		CurrentProxyConnectionCount: int64(s.CurrentProxyConnectionCount),
		ProxyConnectCount:           int64(s.ProxyConnectTotalCount),
		ProxyNetFlowInBytes:         int64(s.NetworkFlowTotalCountInBytes),
	}
	for _, c := range s.ProxyNetworkFlowTop10 {
		res.ConfigNetFlowTop10 = append(res.ConfigNetFlowTop10, &pb.ProxyConfig{
			AgentId:                         c.AgentId,
			RemotePort:                      int64(c.RemotePort),
			LocalAddr:                       c.LocalAddr,
			IsWhiteListOn:                   c.IsWhiteListOn,
			WhiteCidrList:                   c.WhiteCidrList,
			NetworkFlowRemoteToLocalInBytes: int64(c.NetworkFlowRemoteToLocalInBytes),
			NetworkFlowLocalToRemoteInBytes: int64(c.NetworkFlowLocalToRemoteInBytes),
			ProxyConnectCount:               int64(c.ProxyConnectCount),
			ProxyConnectRejectCount:         int64(c.ProxyConnectRejectCount),
		})
	}

	for _, c := range s.ProxyConnectRejectCountTop10 {
		res.ConfigConnectFailTop10 = append(res.ConfigConnectFailTop10, &pb.ProxyConfig{
			AgentId:                         c.AgentId,
			RemotePort:                      int64(c.RemotePort),
			LocalAddr:                       c.LocalAddr,
			IsWhiteListOn:                   c.IsWhiteListOn,
			WhiteCidrList:                   c.WhiteCidrList,
			NetworkFlowRemoteToLocalInBytes: int64(c.NetworkFlowRemoteToLocalInBytes),
			NetworkFlowLocalToRemoteInBytes: int64(c.NetworkFlowLocalToRemoteInBytes),
			ProxyConnectCount:               int64(c.ProxyConnectCount),
			ProxyConnectRejectCount:         int64(c.ProxyConnectRejectCount),
		})
	}
	return res, nil
}

func (h *rpcHandlers) SendReport(ctx context.Context, empty *pb.Empty) (*pb.Empty, error) {
	h.s.SendDailyReport()
	return &pb.Empty{}, nil
}
