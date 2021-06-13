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
			ZoneName:                        config.ZoneName,
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
	if err := serverInst.AddProxyConfig(params.UserName, params.ZoneName, int(params.RemotePort), params.LocalAddr, params.WhiteListEnable, whiteListIps); err != nil {
		return nil, err
	}
	return &models.ProxyConfig{
		UserName:      params.UserName,
		ZoneName:      params.ZoneName,
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
	if err := serverInst.RemoveProxyConfig(params.UserName, params.
		ZoneName, remotePort, params.LocalAddr); err != nil {
		return nil, err
	}
	return &models.ProxyConfig{
		UserName:  params.UserName,
		ZoneName:  params.ZoneName,
		LocalAddr: params.LocalAddr,
	}, nil
}

func UpdateProxyConfigV1(params v1.PostV1ProxyUpdateParams) (*models.ProxyConfig, error) {
	if err := serverInst.UpdateProxyConfigWhiteList(params.UserName, params.ZoneName, int(util.Int64Nvl(params.RemotePort)),
		params.LocalAddr, util.StringNvl(params.WhiteListIps), util.BoolNvl(params.WhiteListEnable)); err != nil {
		return nil, err
	}
	return &models.ProxyConfig{
		UserName:      params.UserName,
		ZoneName:      params.ZoneName,
		IsWhitelistOn: util.BoolNvl(params.WhiteListEnable),
		LocalAddr:     params.LocalAddr,
		WhitelistIps:  util.StringNvl(params.WhiteListIps),
	}, nil
}

func ListZonesV1(params v1.GetV1ZoneListParams) (res []*models.Zone) {
	res = make([]*models.Zone, 0)
	zones := serverInst.ListZones()
	for _, z := range zones {
		res = append(res, &models.Zone{
			UserName:   z.UserName,
			ZoneName:   z.ZoneName,
			AgentCount: z.AgentsCount,
		})
	}
	return
}
