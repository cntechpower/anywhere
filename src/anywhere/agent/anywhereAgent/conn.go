package anywhereAgent

import (
	"anywhere/conn"
	"anywhere/log"
	"anywhere/model"
	"anywhere/tls"
	_tls "crypto/tls"
	"time"
)

func (a *Agent) getTlsConnToServer() *_tls.Conn {
	var c *_tls.Conn
	for {
		var err error
		c, err = tls.DialTlsServer(a.Addr.IP.String(), a.Addr.Port, a.credential)
		if err != nil {
			log.Error("can not connect to server %v, error: %v", a.Addr, err)
			// sleep 5 second and retry
			time.Sleep(5 * time.Second)
			continue
		}
		return c
	}
}

func (a *Agent) connectControlConn() {
	c := a.getTlsConnToServer()
	a.AdminConn = conn.NewBaseConn(c)
	a.status = "RUNNING"
	if err := a.SendControlConnRegisterPkg(); err != nil {
		log.Error("can not send register pkg to server %v, error: %v", a.Addr, err)
	}
}

func (a *Agent) ControlConnHeartBeatLoop(dur int) {
	go func() {
		for {

			//check conn status first
			//if a.AdminConn.GetStatus() == conn.CStatusBad || a.AdminConn.GetFailCount() >= 3 {
			if a.AdminConn.GetFailCount() >= 3 {
				log.Error("control connection not healthy, status: %v, failCount: %v, failReason: %v", a.AdminConn.GetStatus(), a.AdminConn.GetFailCount(), a.AdminConn.GetFailReason())
				a.connectControlConn()
				log.Info("rebuild control connection to server %v, addr %v", a.ServerId, a.Addr)
				//after control conn rebuild, set conn status to healthy
				a.AdminConn.SetHealthy()
			}

			//if conn status is ok ,generate pkg and send
			m := model.NewHeartBeatMsg(a.AdminConn.GetRawConn())
			msg := model.NewRequestMsg(a.version, model.PkgReqHeartBeat, a.Id, "", m)
			err := a.AdminConn.Send(msg)
			if err != nil {
				a.AdminConn.SetBad(err.Error())
			} else {
				a.AdminConn.SetHealthy()
			}
			log.Info("send heartbeat to %v, error: %v", a.Addr, err)
			time.Sleep(time.Duration(dur) * time.Second)
		}
	}()

}

func (a *Agent) connectDataConn() {
	//init 10 data connections
	for i := 0; i < 10; i++ {
		c := a.getTlsConnToServer()
		baseC := conn.NewBaseConn(c)
		m := model.NewDataConnRegisterMsg(a.Id)
		msg := model.NewRequestMsg(a.version, model.PkgDataConnRegister, a.Id, "", m)
		if err := baseC.Send(msg); err != nil {
			i--
			log.Error("init data conn error :%v", err)
			continue
		}
		a.DataConn = append(a.DataConn, baseC)
	}
}
