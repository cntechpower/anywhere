package conn

import (
	"anywhere/log"
	"anywhere/model"
	"time"
)

func HeartBeatRcvLoop(c *BaseConn, funcOnFail func(c *BaseConn)) {
	msg := &model.RequestMsg{}
	go func(c *BaseConn) {
		for {
			select {
			case <-c.StopRcvChan:
				return
			default:
			}
			if err := c.Receive(&msg); err != nil {
				log.Error("receive from data conn %v  error: %v, close this data conn", c.RemoteAddr(), err)
				_ = c.Close()
				funcOnFail(c)
				return
			}
			switch msg.ReqType {
			case model.PkgReqHeartBeat:
				m, _ := model.ParseHeartBeatPkg(msg.Message)
				c.LastAckSendTime = m.SendTime
				c.LastAckRcvTime = time.Now()
				c.SetHealthy()
			case model.PkgDataConnTunnel:
				log.Info("got data conn tunnel, exit handleDataConnection for %v", c.RemoteAddr())
				return
			default:
				log.Error("got unknown ReqType: %v from %v", msg.ReqType, c.RemoteAddr())
				_ = c.Close()
				funcOnFail(c)
				return
			}
		}
	}(c)
}
