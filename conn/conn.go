package conn

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

var ErrNilConn = fmt.Errorf("empty net.Conn")

// encode 将消息编码
func encode(message []byte) ([]byte, error) {
	// 读取消息的长度，转换成int32类型（占4个字节）
	var length = int32(len(message))
	var pkg = new(bytes.Buffer)
	// 写入消息头
	err := binary.Write(pkg, binary.LittleEndian, length)
	if err != nil {
		return nil, err
	}
	// 写入消息实体
	err = binary.Write(pkg, binary.LittleEndian, message)
	if err != nil {
		return nil, err
	}
	return pkg.Bytes(), nil
}

// decode 解码消息
func decode(reader *bufio.Reader) ([]byte, error) {
	// 读取消息的长度
	lengthByte, _ := reader.Peek(4) // 读取前4个字节的数据
	lengthBuff := bytes.NewBuffer(lengthByte)
	var length int32
	err := binary.Read(lengthBuff, binary.LittleEndian, &length)
	if err != nil {
		return nil, err
	}
	// Buffered返回缓冲中现有的可读取的字节数。
	if int32(reader.Buffered()) < length+4 {
		return nil, fmt.Errorf("size not enough")
	}

	// 读取真正的消息数据
	pack := make([]byte, int(4+length))
	_, err = reader.Read(pack)
	if err != nil {
		return nil, err
	}
	return pack[4:], nil
}

type WrappedConn struct {
	RemoteName      string
	connRwMu        sync.RWMutex
	Conn            net.Conn
	statusRwMutex   sync.RWMutex
	CreateTime      time.Time
	LastAckSendTime time.Time
	LastAckRcvTime  time.Time
}

func (c *WrappedConn) SetAck(sendTime, rcvTime time.Time) {
	c.statusRwMutex.Lock()
	defer c.statusRwMutex.Unlock()
	c.LastAckSendTime = sendTime
	c.LastAckRcvTime = rcvTime
}

func (c *WrappedConn) Send(m interface{}) error {
	c.connRwMu.RLock()
	defer c.connRwMu.RUnlock()
	if c.Conn == nil {
		return ErrNilConn
	}
	p, err := json.Marshal(m)
	if err != nil {
		return err
	}
	msg, err := encode(p)
	if err != nil {
		return err
	}
	if _, err := c.Conn.Write(msg); err != nil {
		return err
	}
	return nil
}

func (c *WrappedConn) Receive(rsp interface{}) error {
	c.connRwMu.RLock()
	defer c.connRwMu.RUnlock()
	if c.Conn == nil {
		return ErrNilConn
	}
	reader := bufio.NewReader(c.Conn)
	msg, err := decode(reader)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(msg, &rsp); err != nil {
		return err
	}
	return nil
}

func (c *WrappedConn) Close() error {
	c.connRwMu.Lock()
	defer c.connRwMu.Unlock()
	if c.Conn == nil {
		return nil
	}
	err := c.Conn.Close()

	// set conn to nil because net.Conn do not have a isClose flag.
	// we used conn == nil to validate conn
	c.Conn = nil
	return err
}

func (c *WrappedConn) GetRemoteAddr() string {
	if c.Conn == nil {
		return ""
	}
	return c.Conn.RemoteAddr().String()
}

func (c *WrappedConn) GetLocalAddr() string {
	if c.Conn == nil {
		return ""
	}
	return c.Conn.LocalAddr().String()
}

func (c *WrappedConn) IsValid() bool {
	return c.Conn != nil
}

func (c *WrappedConn) ResetConn(conn net.Conn) {
	// close old connection if existed, let old goroutine stop.
	_ = c.Close()

	c.connRwMu.Lock()
	defer c.connRwMu.Unlock()
	c.Conn = conn
}

func (c *WrappedConn) GetConn() net.Conn {
	return c.Conn
}

func NewWrappedConn(remoteName string, c net.Conn) *WrappedConn {
	return &WrappedConn{
		RemoteName:      remoteName,
		Conn:            c,
		statusRwMutex:   sync.RWMutex{},
		CreateTime:      time.Now(),
		LastAckSendTime: time.Time{},
		LastAckRcvTime:  time.Time{},
	}
}
