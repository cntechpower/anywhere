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
	if scoped, restricted := a.scopedUser(params.HTTPRequest); restricted {
		if userName != "" && userName != scoped {
			return nil, errForbidden
		}
		userName = scoped
	}
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
	if scoped, restricted := a.scopedUser(params.HTTPRequest); restricted {
		groups, listErr := a.server.ListJoinedConns(scoped, "")
		if listErr != nil {
			return nil, listErr
		}
		owned := false
		for _, group := range groups {
			for _, item := range group.List {
				if int64(item.ID) == params.ID {
					owned = true
					break
				}
			}
		}
		if !owned {
			return nil, errForbidden
		}
	}
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
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	records, err := whitelist.GetWhiteListDenyRank(limit)
	if err != nil {
		return
	}
	scoped, restricted := a.scopedUser(params.HTTPRequest)
	for _, r := range records {
		if restricted && r.UserName != scoped {
			continue
		}
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
