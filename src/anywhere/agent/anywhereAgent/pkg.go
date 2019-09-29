package anywhereAgent

import (
	"anywhere/model"
	"encoding/json"
	"fmt"
	"net"
)

func (a *Agent) SendProxyConfig(remotePort, localIp, localPort string) error {
	p, err := model.NewProxyConfigMsg(remotePort, localIp, localPort)
	if err != nil {
		return err
	}
	msg := model.NewRequestMsg(a.version, model.PkgReqNewproxy, a.Id, "", p)
	return a.AdminConn.Send(msg)
}

func (a *Agent) SendHeartBeatPkg(c net.Conn) error {
	p := model.NewHeartBeatMsg(c)
	pkg := model.NewRequestMsg(a.version, model.PkgReqHeartBeat, a.Id, "", p)
	pByte, _ := json.Marshal(pkg)
	_, err := c.Write(pByte)
	return err
}

func (a *Agent) SendRegisterPkg() error {
	if a.Id == "" {
		return fmt.Errorf("agent not init")
	}
	p := model.NewRegisterMsg(a.Id)
	pkg := model.NewRequestMsg(a.version, model.PkgRegister, a.Id, "", p)
	return a.AdminConn.Send(pkg)
}
