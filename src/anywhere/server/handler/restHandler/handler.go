package restHandler

import (
	"anywhere/server/handler/rpcHandler"
	"anywhere/server/restapi/api/models"
	pb "anywhere/server/rpc/definitions"
	"context"
)

func ListAgentV1() ([]*models.AgentListInfo, error) {
	c, err := rpcHandler.NewClient()
	if err != nil {
		return nil, err
	}
	res := make([]*models.AgentListInfo, 0)
	agents, err := c.ListAgent(context.Background(), &pb.Empty{})
	for _, agent := range agents.Agent {
		a := &models.AgentListInfo{
			AgentAdminAddr: agent.AgentRemoteAddr,
			AgentID:        agent.AgentId,
			LastAck:        agent.AgentLastAckRcv,
			Status:         agent.AgentStatus,
		}
		res = append(res, a)
	}
	return res, nil

}
