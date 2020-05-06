package anywhereAgent

import (
	"anywhere/model"
	"fmt"
)

func (a *Agent) sendHeartBeatPkg() error {
	if a.adminConn == nil {
		return fmt.Errorf("admin conn not init")
	}
	return a.adminConn.Send(model.NewHeartBeatPingMsg(a.adminConn, a.id))
}

func (a *Agent) SendControlConnRegisterPkg() error {
	if a.id == "" {
		return fmt.Errorf("agent not init")
	}
	return a.adminConn.Send(model.NewAgentRegisterMsg(a.id))
}
