package anywhereAgent

import (
	"anywhere/model"
	"fmt"
)

func (a *Agent) SendHeartBeatPkg() error {
	if a.AdminConn == nil {
		return fmt.Errorf("admin conn not init")
	}
	p := model.NewHeartBeatMsg(a.AdminConn)
	pkg := model.NewRequestMsg(a.version, model.PkgReqHeartBeat, a.Id, "", p)
	return a.AdminConn.Send(pkg)
}

func (a *Agent) SendControlConnRegisterPkg() error {
	if a.Id == "" {
		return fmt.Errorf("agent not init")
	}
	p := model.NewAgentRegisterMsg(a.Id)
	pkg := model.NewRequestMsg(a.version, model.PkgControlConnRegister, a.Id, "", p)
	return a.AdminConn.Send(pkg)
}
