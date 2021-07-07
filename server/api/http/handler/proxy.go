package handler

import (
	"net/http"

	"github.com/cntechpower/anywhere/server/api/http/api/models"
	v1 "github.com/cntechpower/anywhere/server/api/http/api/restapi/operations"
	"github.com/cntechpower/anywhere/util"
)

func ListProxyV1() ([]*models.ProxyConfig, error) {
	res := make([]*models.ProxyConfig, 0)
	configs := serverInst.ListProxyConfigs()
	for _, config := range configs {
		res = append(res, &models.ProxyConfig{
			ID:                              int64(config.ID),
			UserName:                        config.UserName,
			ZoneName:                        config.ZoneName,
			LocalAddr:                       config.LocalAddr,
			RemotePort:                      int64(config.RemotePort),
			IsWhitelistOn:                   config.IsWhiteListOn,
			WhitelistIps:                    config.WhiteCidrList,
			NetworkFlowLocalToRemoteInBytes: int64(config.NetworkFlowLocalToRemoteInBytes),
			NetworkFlowRemoteToLocalInBytes: int64(config.NetworkFlowRemoteToLocalInBytes),
			ProxyConnectCount:               int64(config.ProxyConnectCount),
			ProxyConnectRejectCount:         int64(config.ProxyConnectRejectCount),
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

func PostV1ProxyDeleteHandler(params v1.PostV1ProxyDeleteParams) (res *models.GenericResponse, err error) {
	res = &models.GenericResponse{}
	err = serverInst.RemoveProxyConfigById(uint(params.ID))
	if err == nil {
		res.Code = http.StatusOK
		res.Message = "OK"
	}
	return
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

func ListZonesV1(_ v1.GetV1ZoneListParams) (res []*models.Zone) {
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
