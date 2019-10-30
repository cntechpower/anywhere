package model

type AgentInfo struct {
	Id         string
	RemoteAddr string
	LastAck    string
	Status     string
}

type ProxyConfigInfo struct {
	AgentId    string
	RemoteAddr string
	LocalAddr  string
}
