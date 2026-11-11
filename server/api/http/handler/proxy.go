package handler

import (
	"net/http"

	"github.com/cntechpower/anywhere/server/api/http/api/models"
	v1 "github.com/cntechpower/anywhere/server/api/http/api/restapi/operations"
	"github.com/cntechpower/anywhere/util"
)

func (a *API) ListProxyV1() ([]*models.ProxyConfig, error) {
	res := make([]*models.ProxyConfig, 0)
	configs := a.server.ListProxyConfigs()
	for _, config := range configs {
		res = append(res, toAPIConfig(config))
	}
	return res, nil
}

func (a *API) AddProxyConfigV1(params v1.PostV1ProxyAddParams) (*models.ProxyConfig, error) {
	whiteListIps := util.StringNvl(params.WhiteListIps)
	if err := a.server.AddProxyConfig(
		params.UserName, params.ZoneName, int(params.RemotePort), params.LocalAddr,
		params.WhiteListEnable, whiteListIps, util.StringNvl(params.ListenType),
	); err != nil {
		return nil, err
	}
	return &models.ProxyConfig{
		UserName:      params.UserName,
		ZoneName:      params.ZoneName,
		IsWhitelistOn: params.WhiteListEnable,
		LocalAddr:     params.LocalAddr,
		RemotePort:    params.RemotePort,
		WhitelistIps:  util.StringNvl(params.WhiteListIps),
		ListenType:    util.StringNvl(params.ListenType),
	}, nil
}

func (a *API) PostV1ProxyDeleteHandler(params v1.PostV1ProxyDeleteParams) (res *models.GenericResponse, err error) {
	res = &models.GenericResponse{}
	err = a.server.RemoveProxyConfigById(uint(params.ID))
	if err == nil {
		res.Code = http.StatusOK
		res.Message = "OK"
	}
	return
}

func (a *API) UpdateProxyConfigV1(params v1.PostV1ProxyUpdateParams) (*models.ProxyConfig, error) {
	if err := a.server.UpdateProxyConfigWhiteList(
		params.UserName, params.ZoneName, int(util.Int64Nvl(params.RemotePort)),
		params.LocalAddr, util.StringNvl(params.WhiteListIps), util.BoolNvl(params.WhiteListEnable),
	); err != nil {
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

func (a *API) ListZonesV1(_ v1.GetV1ZoneListParams) []*models.Zone {
	res := make([]*models.Zone, 0)
	for _, z := range a.server.ListZones() {
		res = append(res, &models.Zone{
			UserName:   z.UserName,
			ZoneName:   z.ZoneName,
			AgentCount: z.AgentsCount,
		})
	}
	return res
}
