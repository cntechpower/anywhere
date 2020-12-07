package http

import (
	"github.com/cntechpower/anywhere/server/restapi/api/models"
)

func ListAgentV1() ([]*models.AgentListInfo, error) {
	res := make([]*models.AgentListInfo, 0)
	agents := serverInst.ListAgentInfo()
	for _, agent := range agents {
		a := &models.AgentListInfo{
			UserName:         agent.UserName,
			AgentAdminAddr:   agent.RemoteAddr,
			AgentID:          agent.Id,
			LastAckSend:      agent.LastAckSend,
			LastAckRcv:       agent.LastAckRcv,
			ProxyConfigCount: int64(agent.ProxyConfigCount),
		}
		res = append(res, a)
	}
	return res, nil
}
