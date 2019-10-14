package anywhereAgent

import (
	"anywhere/model"
	"fmt"
)

func (a *Agent) SendProxyConfig(remotePort, localIp, localPort string) error {
	if a.AdminConn == nil {
		return fmt.Errorf("admin conn not init")
	}
	p, err := model.NewProxyConfigMsg(remotePort, localIp, localPort)
	if err != nil {
		return err
	}
	msg := model.NewRequestMsg(a.version, model.PkgReqNewproxy, a.Id, "", p)
	return a.AdminConn.Send(msg)
}

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
