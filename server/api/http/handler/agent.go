package handler

import (
	"github.com/cntechpower/anywhere/constants"
	"github.com/cntechpower/anywhere/server/api/http/api/models"
)

func (a *API) ListAgentV1() ([]*models.AgentListInfo, error) {
	res := make([]*models.AgentListInfo, 0)
	agents := a.server.ListAgentInfo()
	for _, agent := range agents {
		info := &models.AgentListInfo{
			UserName:         agent.UserName,
			ZoneName:         agent.ZoneName,
			AgentAdminAddr:   agent.RemoteAddr,
			AgentID:          agent.Id,
			LastAckSend:      agent.LastAckSend.Local().Format(constants.DefaultTimeFormat),
			LastAckRcv:       agent.LastAckRcv.Local().Format(constants.DefaultTimeFormat),
			ProxyConfigCount: int64(agent.ProxyConfigCount),
		}
		res = append(res, info)
	}
	return res, nil
}
