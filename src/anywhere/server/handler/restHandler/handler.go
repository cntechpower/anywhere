package restHandler

import (
	"anywhere/server/handler/rpcHandler"
	"anywhere/server/restapi/api/models"
	v1 "anywhere/server/restapi/api/restapi/operations"
	pb "anywhere/server/rpc/definitions"
	"anywhere/util"

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
			RemotePort:    config.RemotePort,
			IsWhitelistOn: config.IsWhiteListOn,
			WhitelistIps:  config.WhiteCidrList,
		})
	}
	return res, nil
}

func AddProxyConfigV1(params v1.PostV1ProxyAddParams) (*models.ProxyConfigInfo, error) {
	c, err := rpcHandler.NewClient()
	if err != nil {
		return nil, err
	}
	if _, err := c.AddProxyConfig(context.Background(), &pb.AddProxyConfigInput{
		Config: &pb.ProxyConfig{
			AgentId:       params.AgentID,
			RemotePort:    params.RemotePort,
			LocalAddr:     params.LocalAddr,
			IsWhiteListOn: params.WhiteListEnable,
			WhiteCidrList: util.StringNvl(params.WhiteListIps),
		},
	}); err != nil {
		return nil, err
	}
	return &models.ProxyConfigInfo{
		AgentID:       params.AgentID,
		IsWhitelistOn: params.WhiteListEnable,
		LocalAddr:     params.LocalAddr,
		RemotePort:    params.RemotePort,
		WhitelistIps:  util.StringNvl(params.WhiteListIps),
	}, nil

}
