package handler

import (
	"net/http"
	"time"

	"github.com/cntechpower/anywhere/server/api/http/api/models"
)

func (a *API) ListAgentV1(r *http.Request) ([]*models.AgentListInfo, error) {
	res := make([]*models.AgentListInfo, 0)
	scoped, restricted := a.scopedUser(r)
	agents := a.server.ListAgentInfo()
	for _, agent := range agents {
		if restricted && agent.UserName != scoped {
			continue
		}
		info := &models.AgentListInfo{
			UserName:         agent.UserName,
			ZoneName:         agent.ZoneName,
			AgentAdminAddr:   agent.RemoteAddr,
			AgentID:          agent.Id,
			LastAckSend:      agent.LastAckSend.UTC().Format(time.RFC3339),
			LastAckRcv:       agent.LastAckRcv.UTC().Format(time.RFC3339),
			ProxyConfigCount: int64(agent.ProxyConfigCount),
		}
		res = append(res, info)
	}
	return res, nil
}
