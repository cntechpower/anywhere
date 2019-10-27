package handler

import (
	"anywhere/server/restapi/api/models"
	pb "anywhere/server/rpc/definitions"
	"context"
	"fmt"
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

func (h *rpcHandlers) ListAgent(ctx context.Context, empty *pb.Empty) (*pb.AgentInfo, error) {
	s := anywhereServer.GetServerInstance()
	if s == nil {
		return nil, fmt.Errorf("anywhere server not init")
	}
	res := &pb.AgentInfo{}
	agents := s.ListAgentInfoStruct()
	for _, agent := range agents {
		res.AgentId = agent.Id
	}
	return res, nil
}
