package handler

import (
	"net/http"

	"github.com/cntechpower/anywhere/dao/whitelist"

	"github.com/cntechpower/anywhere/server/api/http/api/models"
	v1 "github.com/cntechpower/anywhere/server/api/http/api/restapi/operations"
	"github.com/cntechpower/anywhere/util"
)

func (a *API) GetConnsV1(params v1.GetV1ConnectionListParams) (res []*models.ConnListItem, err error) {
	res = make([]*models.ConnListItem, 0)
	userName := util.StringNvl(params.UserName)
	zoneName := util.StringNvl(params.ZoneName)
	groups, err := a.server.ListJoinedConns(userName, zoneName)
	if err != nil {
		return
	}
	for _, group := range groups {
		for _, l := range group.List {
			res = append(res, &models.ConnListItem{
				ID:            int64(l.ID),
				DstLocalAddr:  l.DstLocalAddr,
				DstName:       l.DstName,
				DstRemoteAddr: l.DstRemoteAddr,
				SrcLocalAddr:  l.SrcLocalAddr,
				SrcName:       l.SrcName,
				SrcRemoteAddr: l.SrcRemoteAddr,
			})
		}
	}
	return
}

func (a *API) KillConnV1(params v1.PostV1ConnectionKillParams) (res *models.GenericResponse, err error) {
	res = &models.GenericResponse{}
	err = a.server.KillJoinedConnById(params.ID)
	if err == nil {
		res.Code = http.StatusOK
		res.Message = "OK"
	}
	return
}

func (a *API) WhiteListRecordV1(params v1.GetV1WhitelistDenysParams) (res []*models.WhiteListDenyRecordItem, err error) {
	res = make([]*models.WhiteListDenyRecordItem, 0)
	limit := int(util.Int64Nvl(params.Limit))
	records, err := whitelist.GetWhiteListDenyRank(limit)
	if err != nil {
		return
	}
	for _, r := range records {
		res = append(res, &models.WhiteListDenyRecordItem{
			Ctime:     r.CreatedAt.Unix(),
			ID:        int64(r.ID),
			IP:        r.IP,
			LocalAddr: r.LocalAddr,
			UserName:  r.UserName,
			ZoneName:  r.ZoneName,
		})
	}
	return
}
