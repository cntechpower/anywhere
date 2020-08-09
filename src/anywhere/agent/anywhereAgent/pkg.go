package anywhereAgent

import (
	"anywhere/model"
	"fmt"
)

func (a *Agent) sendHeartBeatPkg() error {
	if !a.adminConn.IsValid() {
		return fmt.Errorf("admin conn not init")
	}
	return a.adminConn.Send(model.NewHeartBeatPingMsg(a.adminConn.GetLocalAddr(), a.adminConn.GetRemoteAddr(), a.id))
}

func (a *Agent) SendControlConnRegisterPkg() error {
	if a.id == "" {
		return fmt.Errorf("agent not init")
	}
	return a.adminConn.Send(model.NewAgentRegisterMsg(a.id, a.user, a.password))
}
