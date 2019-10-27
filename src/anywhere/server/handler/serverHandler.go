package handler

import (
	"anywhere/server/restapi/api/models"
	pb "anywhere/server/rpc/definitions"
	"context"
	"fmt"
	"strconv"
)
import "anywhere/server/anywhereServer"

func ListAgentV1() ([]*models.AgentListInfo, error) {
	res := make([]*models.AgentListInfo, 0)
	s := anywhereServer.GetServerInstance()
	if s == nil {
		return nil, fmt.Errorf("anywhere server not init")
	}
	agents := s.ListAgentInfoStruct()
	for _, agent := range agents {
		a := &models.AgentListInfo{
			AgentAdminAddr: agent.AdminConn.RemoteAddr().String(),
			AgentID:        agent.Id,
			LastAck:        "",
			Status:         "Healthy",
		}
		res = append(res, a)
	}
	return res, nil

}

type rpcHandlers struct {
}

func GetRpcHandlers() *rpcHandlers {
	return &rpcHandlers{}
}

func (h *rpcHandlers) ListAgent(ctx context.Context, empty *pb.Empty) (*pb.Agents, error) {
	s := anywhereServer.GetServerInstance()
	if s == nil {
		return nil, fmt.Errorf("anywhere server not init")
	}
	res := &pb.Agents{
		Agent: make([]*pb.Agent, 0),
	}
	agents := s.ListAgentInfoStruct()
	for _, agent := range agents {
		res.Agent = append(res.Agent, &pb.Agent{
			AgentId:         agent.Id,
			AgentVersion:    "",
			AgentRemoteAddr: agent.RemoteAddr.String(),
		})
	}
	return res, nil
}

func (h *rpcHandlers) AddProxyConfig(ctx context.Context, input *pb.AddProxyConfigInput) (*pb.Empty, error) {
	s := anywhereServer.GetServerInstance()
	if s == nil {
		return nil, fmt.Errorf("anywhere server not init")
	}
	remotePort, err := strconv.Atoi(input.RemotePort)
	if err != nil {
		return nil, err
	}
	localPort, err := strconv.Atoi(input.LocalPort)
	if err != nil {
		return nil, err
	}
	if err := s.AddProxyConfigToAgent(input.AgentId, remotePort, input.LocalIp, localPort); err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}
