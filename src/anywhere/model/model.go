package model

type AgentInfo struct {
	Id          string
	RemoteAddr  string
	LastAckSend string
	LastAckRcv  string
}

type ProxyConfigInfo struct {
	AgentId    string
	RemoteAddr string
	LocalAddr  string
}
