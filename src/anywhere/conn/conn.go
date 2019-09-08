package conn

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

func ListenAndServeTls(port int, config *tls.Config) error {
	addr := fmt.Sprintf("0.0.0.0:%v", port)
	ln, err := tls.Listen("tcp", addr, config)
	if err != nil {
		return err
	}
	go func(ln net.Listener) {
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Printf("accept conn error: %v", err)
				continue
			}
			if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
				fmt.Printf("set readtimeout error: %v", err)
			}
			go handleConnection(conn)
		}
	}(ln)
	return nil
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	d := json.NewDecoder(conn)
	var msg Package
	if err := d.Decode(&msg); err != nil {
		fmt.Printf("decode error :%v\n", err)
		if _, err := conn.Write([]byte("invalid message\n")); err != nil {
			fmt.Printf("error while send invalid message warning to client: %v\n", err)
		}
		if err := conn.Close(); err != nil {
			fmt.Printf("error while close connection: %v\n", err)
		}
	} else {
		fmt.Println(msg)
		p, err := json.Marshal(&Package{
			Version: "3.19.09.0",
			Type:    "admin",
			Message: "HelloWorldFromServer",
		})
		if err != nil {
			panic(err)
		}
		if _, err := conn.Write(p); err != nil {
			fmt.Printf("send message error: %v", err)
		}

	}
	time.Sleep(1 * time.Millisecond)
}

func GetDialer() *net.Dialer {
	return &net.Dialer{
		Timeout:  5 * time.Second,
		Deadline: time.Time{},
	}
}
