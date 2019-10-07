package anywhereServer

import (
	"anywhere/conn"
	"anywhere/model"
	"net"
)

type Agent struct {
	Id               string
	ServerId         string
	RemoteAddr       net.Addr
	AdminConn        *conn.BaseConn
	DataConn         []*DataConn
	ProxyConfigs     []model.ProxyConfig
	chanProxyConfigs chan model.ProxyConfig
}

type DataConn struct {
	*conn.BaseConn
	InUsed bool
}

func NewAgentInfo(agentId, serverId string, c net.Conn) *Agent {
	return &Agent{
		Id:         agentId,
		ServerId:   serverId,
		RemoteAddr: c.RemoteAddr(),
		AdminConn:  conn.NewBaseConn(c),
		DataConn:   make([]*DataConn, 0),
	}
}
