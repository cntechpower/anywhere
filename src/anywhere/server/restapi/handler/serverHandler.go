package handler

import (
	"anywhere/server/restapi/api/models"
	"fmt"
)
import "anywhere/server/anywhereServer"

func ListAgentV1() ([]*models.AgentListInfo, error) {
	res := make([]*models.AgentListInfo, 0)
	s := anywhereServer.GetServerInstance()
	if s == nil {
		return nil, fmt.Errorf("anywhere server not init")
	}
	agents := s.ListAgentInfoStruct()
	for _, agent := range agents {
		a := &models.AgentListInfo{
			AgentAdminAddr: agent.AdminConn.RemoteAddr().String(),
			AgentID:        agent.Id,
			LastAck:        "",
			Status:         "Healthy",
		}
		res = append(res, a)
	}
	return res, nil

}
