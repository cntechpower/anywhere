package conn

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

var (
	ErrNilConn       = errors.New("empty net.Conn")
	ErrFrameTooLarge = errors.New("frame too large")
)

const (
	// frameHeaderSize is the size of the length prefix of a frame.
	frameHeaderSize = 4
	// MaxFrameSize is the maximum allowed payload of a single frame (16MiB).
	// It protects the receiver from allocating an arbitrary amount of memory
	// based on a corrupted or malicious length field.
	MaxFrameSize = 16 << 20
)

// writeAll writes b completely, retrying short writes.
func writeAll(w io.Writer, b []byte) error {
	for len(b) > 0 {
		n, err := w.Write(b)
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrUnexpectedEOF
		}
		b = b[n:]
	}
	return nil
}

// writeFrame encodes payload as [4-byte little-endian length][payload].
func writeFrame(w io.Writer, payload []byte) error {
	if len(payload) > MaxFrameSize {
		return fmt.Errorf("%w: %d > %d", ErrFrameTooLarge, len(payload), MaxFrameSize)
	}
	hdr := make([]byte, frameHeaderSize)
	length := uint32(len(payload))
	hdr[0] = byte(length)
	hdr[1] = byte(length >> 8)
	hdr[2] = byte(length >> 16)
	hdr[3] = byte(length >> 24)
	if err := writeAll(w, hdr); err != nil {
		return err
	}
	if len(payload) == 0 {
		return nil
	}
	return writeAll(w, payload)
}

// readFrame reads one frame from r. It blocks until a full frame is available,
// so TCP segmentation and partial reads are handled transparently.
func readFrame(r io.Reader) ([]byte, error) {
	hdr := make([]byte, frameHeaderSize)
	if _, err := io.ReadFull(r, hdr); err != nil {
		return nil, err
	}
	length := uint32(hdr[0]) | uint32(hdr[1])<<8 | uint32(hdr[2])<<16 | uint32(hdr[3])<<24
	if length == 0 {
		return []byte{}, nil
	}
	if length > MaxFrameSize {
		return nil, fmt.Errorf("%w: %d > %d", ErrFrameTooLarge, length, MaxFrameSize)
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

// WrappedConn is a concurrency-safe, message oriented wrapper around net.Conn.
//
// Sending and receiving use independent locks, so one goroutine may write while
// another blocks in Receive. The underlying connection may be replaced with
// ResetConn (used by the agent control channel to reconnect) without leaking
// partially written frames.
type WrappedConn struct {
	RemoteName string

	mu   sync.RWMutex
	conn net.Conn

	// sendMu serializes whole-frame writes so concurrent Send calls cannot
	// interleave their bytes.
	sendMu sync.Mutex
	// recvMu serializes Receive calls. A connection must have a single reader.
	recvMu sync.Mutex

	statusRwMutex   sync.RWMutex
	CreateTime      time.Time
	LastAckSendTime time.Time
	LastAckRcvTime  time.Time
}

func NewWrappedConn(remoteName string, c net.Conn) *WrappedConn {
	return &WrappedConn{
		RemoteName: remoteName,
		conn:       c,
		CreateTime: time.Now(),
	}
}

func (c *WrappedConn) Send(m interface{}) error {
	p, err := json.Marshal(m)
	if err != nil {
		return err
	}
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	nc := c.GetConn()
	if nc == nil {
		return ErrNilConn
	}
	return writeFrame(nc, p)
}

func (c *WrappedConn) Receive(rsp interface{}) error {
	nc := c.GetConn()
	if nc == nil {
		return ErrNilConn
	}
	c.recvMu.Lock()
	defer c.recvMu.Unlock()
	// Re-check after acquiring the receive lock: the connection may have been
	// closed or replaced while we were waiting.
	nc = c.GetConn()
	if nc == nil {
		return ErrNilConn
	}
	payload, err := readFrame(nc)
	if err != nil {
		return err
	}
	return json.Unmarshal(payload, rsp)
}

func (c *WrappedConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return nil
	}
	err := c.conn.Close()
	c.conn = nil
	return err
}

// ResetConn closes the current connection (if any) and installs a new one.
func (c *WrappedConn) ResetConn(nc net.Conn) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		_ = c.conn.Close()
	}
	c.conn = nc
}

func (c *WrappedConn) GetRemoteAddr() string {
	nc := c.GetConn()
	if nc == nil || nc.RemoteAddr() == nil {
		return ""
	}
	return nc.RemoteAddr().String()
}

func (c *WrappedConn) GetLocalAddr() string {
	nc := c.GetConn()
	if nc == nil || nc.LocalAddr() == nil {
		return ""
	}
	return nc.LocalAddr().String()
}

func (c *WrappedConn) IsValid() bool {
	return c.GetConn() != nil
}

// GetConn returns the underlying connection. Callers must not close it
// directly; use Close/ResetConn so the wrapper stays consistent.
func (c *WrappedConn) GetConn() net.Conn {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn
}

func (c *WrappedConn) SetAck(sendTime, rcvTime time.Time) {
	c.statusRwMutex.Lock()
	defer c.statusRwMutex.Unlock()
	c.LastAckSendTime = sendTime
	c.LastAckRcvTime = rcvTime
}

func (c *WrappedConn) GetAckTimes() (sendTime, rcvTime time.Time) {
	c.statusRwMutex.RLock()
	defer c.statusRwMutex.RUnlock()
	return c.LastAckSendTime, c.LastAckRcvTime
}

func (c *WrappedConn) GetLastAckRcvTime() time.Time {
	_, rcv := c.GetAckTimes()
	return rcv
}

func (c *WrappedConn) GetLastAckSendTime() time.Time {
	send, _ := c.GetAckTimes()
	return send
}
