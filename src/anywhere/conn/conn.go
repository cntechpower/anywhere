package conn

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type Listener struct {
	ListenerType    string
	ListenAddress   net.Addr
	ConnectionsChan chan *net.Conn
	Connections     map[string]*net.Conn
}

func ListenAndServeTls(port int, config *tls.Config) error {
	addr := fmt.Sprintf("0.0.0.0:%v", port)
	ln, err := tls.Listen("tcp", addr, config)
	if err != nil {
		return err
	}
	listener := &Listener{
		ListenerType:    "test",
		ListenAddress:   ln.Addr(),
		ConnectionsChan: make(chan *net.Conn, 0),
		Connections:     make(map[string]*net.Conn),
	}
	go func(ln net.Listener, l *Listener) {
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Printf("accept conn error: %v", err)
				continue
			}
			if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
				fmt.Printf("set readtimeout error: %v", err)
			}
			l.Connections[conn.RemoteAddr().String()] = &conn
			l.ConnectionsChan <- &conn
		}
	}(ln, listener)
	go handleConnection(listener)
	return nil
}

func handleConnection(listener *Listener) {
	for conn := range listener.ConnectionsChan {
		go func(c net.Conn) {
			d := json.NewDecoder(c)
			var msg Package
			if err := d.Decode(&msg); err != nil {
				fmt.Println("Decode Package Error")
			}
			fmt.Println(msg)
			if err := c.Close(); err != nil {
				fmt.Printf("Error Close Conn: %v\n", err)
			}
		}(*conn)
	}

}

func GetDialer() *net.Dialer {
	return &net.Dialer{
		Timeout:  5 * time.Second,
		Deadline: time.Time{},
	}
}
