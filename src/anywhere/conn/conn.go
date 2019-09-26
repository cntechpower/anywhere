package conn

import (
	"anywhere/model"
	"encoding/json"
	"net"
)

func SendRequest(c net.Conn, v, t, m string) error {
	p, err := json.Marshal(&model.RequestMsg{
		Version: v,
		ReqType: t,
		Message: m,
	})
	if err != nil {
		return err
	}
	if _, err := c.Write(p); err != nil {
		return err
	}
	return nil
}

func ReadRequest(c net.Conn) (model.RequestMsg, error) {
	d := json.NewDecoder(c)
	var msg model.RequestMsg
	if err := d.Decode(&msg); err != nil {
		return msg, err
	}
	return msg, nil

}

func SendResponse(c net.Conn, code int, m string) error {
	p, err := json.Marshal(&model.ResponseMsg{
		Code:    code,
		Message: m,
	})
	if err != nil {
		return err
	}
	if _, err := c.Write(p); err != nil {
		return err
	}
	return nil
}

func ReadResponse(c net.Conn) (model.ResponseMsg, error) {
	d := json.NewDecoder(c)
	var rsp model.ResponseMsg
	if err := d.Decode(&rsp); err != nil {
		return rsp, err
	}
	return rsp, nil
}
