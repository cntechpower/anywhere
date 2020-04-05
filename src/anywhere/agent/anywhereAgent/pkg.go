package anywhereAgent

import (
	"anywhere/model"
	"fmt"
)

func (a *Agent) sendHeartBeatPkg() error {
	if a.adminConn == nil {
		return fmt.Errorf("admin conn not init")
	}
	p := model.NewHeartBeatMsg(a.adminConn)
	pkg := model.NewRequestMsg(a.version, model.PkgReqHeartBeat, a.id, "", p)
	return a.adminConn.Send(pkg)
}

func (a *Agent) SendControlConnRegisterPkg() error {
	if a.id == "" {
		return fmt.Errorf("agent not init")
	}
	p := model.NewAgentRegisterMsg(a.id)
	pkg := model.NewRequestMsg(a.version, model.PkgControlConnRegister, a.id, "", p)
	return a.adminConn.Send(pkg)
}
