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
	if err != nil {
		return nil, err
	}
	for _, agent := range agents.Agent {
		a := &models.AgentListInfo{
			AgentAdminAddr: agent.AgentRemoteAddr,
			AgentID:        agent.AgentId,
			LastAckSend:    agent.AgentLastAckSend,
			LastAckRcv:     agent.AgentLastAckRcv,
		}
		res = append(res, a)
	}
	return res, nil

}

func ListProxyV1() ([]*models.ProxyConfigInfo, error) {
	c, err := rpcHandler.NewClient()
	if err != nil {
		return nil, err
	}
	res := make([]*models.ProxyConfigInfo, 0)
	configs, err := c.ListProxyConfigs(context.Background(), &pb.Empty{})
	if err != nil {
		return nil, err
	}
	for _, config := range configs.Config {
		res = append(res, &models.ProxyConfigInfo{
			AgentID:       config.AgentId,
			LocalAddr:     config.LocalAddr,
			RemoteAddr:    config.RemoteAddr,
			IsWhitelistOn: config.IsWhiteListOn,
			WhitelistIps:  config.WhiteListIps,
		})
	}
	return res, nil
}
