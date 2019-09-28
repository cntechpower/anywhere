package conn

import "net"

type DataConnList struct {
	connList []*baseConn
}

func (c *DataConnList) HeartBeatLoop(f func(c net.Conn) error) {
	for _, c1 := range c.connList {
		c1.HeartBeatLoop(f)
	}
}
