package handler

import (
	"net/http"

	"github.com/cntechpower/anywhere/dao/connlist"
	"github.com/cntechpower/anywhere/server/api/http/api/models"
	v1 "github.com/cntechpower/anywhere/server/api/http/api/restapi/operations"
	"github.com/cntechpower/anywhere/util"
)

func GetConnsV1(params v1.GetV1ConnectionListParams) (res []*models.ConnListItem, err error) {
	res = make([]*models.ConnListItem, 0)
	userName := util.StringNvl(params.UserName)
	zoneName := util.StringNvl(params.ZoneName)
	list, err := connlist.GetJoinedConnList(userName, zoneName)
	if err != nil {
		return
	}
	for _, l := range list {
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
	return
}

func KillConnV1(params v1.PostV1ConnectionKillParams) (res *models.GenericResponse, err error) {
	res = &models.GenericResponse{}
	err = serverInst.KillJoinedConnById(params.ID)
	if err == nil {
		res.Code = http.StatusOK
		res.Message = "OK"
	}
	return
}
