package anywhereServer

import (
	"anywhere/conn"
	"net"
)

type AgentInfo struct {
	Id         string
	ServerId   string
	RemoteAddr net.Addr
	AdminConn  *conn.AdminConn
	DataConn   []net.Conn
}

func NewAgentInfo(agentId, serverId string, c net.Conn) *AgentInfo {
	return &AgentInfo{
		Id:         agentId,
		ServerId:   serverId,
		RemoteAddr: c.RemoteAddr(),
		AdminConn:  conn.NewAdminConn(c),
		DataConn:   nil,
	}
}
