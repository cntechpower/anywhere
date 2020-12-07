package http

import (
	"strconv"

	"github.com/cntechpower/anywhere/server/restapi/api/models"
	v1 "github.com/cntechpower/anywhere/server/restapi/api/restapi/operations"
	"github.com/cntechpower/anywhere/util"
)

func ListProxyV1() ([]*models.ProxyConfig, error) {
	res := make([]*models.ProxyConfig, 0)
	configs := serverInst.ListProxyConfigs()
	for _, config := range configs {
		res = append(res, &models.ProxyConfig{
			UserName:                        config.UserName,
			AgentID:                         config.AgentId,
			LocalAddr:                       config.LocalAddr,
			RemotePort:                      int64(config.RemotePort),
			IsWhitelistOn:                   config.IsWhiteListOn,
			WhitelistIps:                    config.WhiteCidrList,
			NetworkFlowLocalToRemoteInBytes: int64(config.NetworkFlowLocalToRemoteInBytes),
			NetworkFlowRemoteToLocalInBytes: int64(config.NetworkFlowRemoteToLocalInBytes),
		})
	}
	return res, nil
}

func AddProxyConfigV1(params v1.PostV1ProxyAddParams) (*models.ProxyConfig, error) {
	whiteListIps := util.StringNvl(params.WhiteListIps)
	if err := serverInst.AddProxyConfigToAgent(params.UserName, params.AgentID, int(params.RemotePort), params.LocalAddr, params.WhiteListEnable, whiteListIps); err != nil {
		return nil, err
	}
	return &models.ProxyConfig{
		UserName:      params.UserName,
		AgentID:       params.AgentID,
		IsWhitelistOn: params.WhiteListEnable,
		LocalAddr:     params.LocalAddr,
		RemotePort:    params.RemotePort,
		WhitelistIps:  util.StringNvl(params.WhiteListIps),
	}, nil
}

func PostV1ProxyDeleteHandler(params v1.PostV1ProxyDeleteParams) (*models.ProxyConfig, error) {
	remotePort, err := strconv.Atoi(params.RemotePort)
	if err != nil {
		return nil, err
	}
	if err := serverInst.RemoveProxyConfigFromAgent(params.UserName, remotePort, params.AgentID, params.LocalAddr); err != nil {
		return nil, err
	}
	return &models.ProxyConfig{
		UserName:  params.UserName,
		AgentID:   params.AgentID,
		LocalAddr: params.LocalAddr,
	}, nil
}

func PostV1ProxyUpdateParams(params v1.PostV1ProxyUpdateParams) (*models.ProxyConfig, error) {
	if err := serverInst.UpdateProxyConfigWhiteList(params.UserName, int(util.Int64Nvl(params.RemotePort)),
		params.AgentID, params.LocalAddr, util.StringNvl(params.WhiteListIps), params.WhiteListEnable); err != nil {
		return nil, err
	}
	return &models.ProxyConfig{
		UserName:      params.UserName,
		AgentID:       params.AgentID,
		IsWhitelistOn: params.WhiteListEnable,
		LocalAddr:     params.LocalAddr,
		WhitelistIps:  util.StringNvl(params.WhiteListIps),
	}, nil
}
