package anywhereAgent

import (
	"anywhere/conn"
	"anywhere/log"
	"anywhere/model"
	"anywhere/tls"
	"context"
	_tls "crypto/tls"
	"encoding/json"
	"net"
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
			if a.AdminConn.GetStatus() == conn.CStatusBad {
				log.Error("control connection not healthy, status: %v, failReason: %v", a.AdminConn.GetStatus(), a.AdminConn.GetFailReason())
				a.connectControlConn()
				log.Info("rebuild control connection to server %v, addr %v", a.ServerId, a.Addr)
				//after control conn rebuild, set conn status to healthy
				a.AdminConn.SetHealthy()
			}

			//if conn status is ok ,generate pkg and send
			m := model.NewHeartBeatMsg(a.AdminConn)
			msg := model.NewRequestMsg(a.version, model.PkgReqHeartBeat, a.Id, "", m)
			err := a.AdminConn.Send(msg)
			if err != nil {
				a.AdminConn.SetBad(err.Error())
			} else {
				a.AdminConn.SetHealthy()
			}
			time.Sleep(time.Duration(dur) * time.Second)
		}
	}()

}

func (a *Agent) connectDataConn(count int) {
	//init 10 data connections
	for i := 0; i < count; i++ {
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

func (a *Agent) dataConnHeartBeatSendLoop(dur int) {

	for idx, c := range a.DataConn {
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		c.CancelHeartBeatSend = cancel
		go func(idx int, c *conn.BaseConn, ctx context.Context) {
			defer log.Error("data conn heartbeat loop for %v exit", c.LocalAddr())
			for {
				select {
				case <-c.StopSendChan:
					return
				default:
				}
				if c.GetStatus() == conn.CStatusBad {
					log.Error("data connection not healthy, status: %v, failReason: %v", c.GetStatus(), c.GetFailReason())
					a.DataConn = append(a.DataConn[:idx], a.DataConn[idx+1:]...)
					_ = c.Close()
					a.connectDataConn(1)
					log.Info("rebuild data connection to server %v, addr %v", a.ServerId, a.Addr)
				}
				//if conn status is ok ,generate pkg and send
				m := model.NewHeartBeatMsg(a.AdminConn)
				msg := model.NewRequestMsg(a.version, model.PkgReqHeartBeat, a.Id, "", m)
				err := c.Send(msg)
				if err != nil {
					c.SetBad(err.Error())
				} else {
					c.SetHealthy()
				}
				time.Sleep(time.Duration(dur) * time.Second)
			}

		}(idx, c, ctx)
	}
}

func (a *Agent) dataConnTunnelWatchLoop(dur int) {
	go func() {

		for idx, c := range a.DataConn {
			go func(idx int, c *conn.BaseConn) {
				defer log.Error("tunnel watch loop for %v exit", c.LocalAddr())

				msg := &model.RequestMsg{}
				for {
					select {
					case <-c.StopRcvChan:
						return
					default:
					}
					if err := c.Receive(&msg); err != nil {
						log.Error("receive from data conn error: %v, close this data conn", err)
						_ = c.Close()
						return
					}
					switch msg.ReqType {
					case model.PkgDataConnTunnel:
						m, _ := model.ParseTunnelBeginPkg(msg.Message)
						log.Info("got tunnel request: %v", m.LocalAddr)
						lc, err := net.Dial("tcp", m.LocalAddr)
						if err != nil {
							log.Error("dial local addr %v error: %v", m.LocalAddr, err)
						} else {
							//c.CancelHeartBeatSend()
							//c.CancelHeartBeatReceive()
							c.StopSendChan <- struct{}{}
							p := model.NewTunnelBeginMsg("")
							pkg := model.NewRequestMsg("0.0.1", model.PkgDataConnTunnel, a.Id, "", p)
							pByte, _ := json.Marshal(pkg)
							if _, err = c.Write(pByte); err != nil {
								log.Info("response to server tunnel pkg error: %v", err)
							}
							//c.StopRcvChan <- struct{}{}
							conn.JoinConn(lc, c)
						}

					default:
						log.Error("got unknown ReqType: %v", msg.ReqType)
						_ = c.Close()

					}
				}
			}(idx, c)

		}
	}()
}
