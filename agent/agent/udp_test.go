package agent

import (
	"net"
	"testing"
	"time"

	"github.com/cntechpower/anywhere/model"
)

// TestUDPManagerForwardsAndReturns verifies the full agent side UDP path:
// datagram -> local service -> reply -> PkgUDPDataResp back to the server.
func TestUDPManagerForwardsAndReturns(t *testing.T) {
	echo, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1")})
	if err != nil {
		t.Fatal(err)
	}
	defer echo.Close()
	go func() {
		buf := make([]byte, 2048)
		for {
			n, addr, err := echo.ReadFromUDP(buf)
			if err != nil {
				return
			}
			_, _ = echo.WriteToUDP(buf[:n], addr)
		}
	}()

	responses := make(chan *model.UDPDataMsg, 4)
	m := newUDPManager()
	m.send = func(msg *model.RequestMsg) error {
		parsed, err := model.ParseUDPDataMsg(msg.Message)
		if err != nil {
			return err
		}
		responses <- parsed
		return nil
	}
	defer m.CloseAll()

	m.Handle(&model.UDPDataMsg{
		SessionID: 7,
		LocalAddr: echo.LocalAddr().String(),
		Data:      []byte("ping"),
	})

	select {
	case got := <-responses:
		if got.SessionID != 7 {
			t.Fatalf("session id = %d, want 7", got.SessionID)
		}
		if string(got.Data) != "ping" {
			t.Fatalf("data = %q, want ping", string(got.Data))
		}
	case <-time.After(3 * time.Second):
		t.Fatal("agent did not forward the udp reply to the server")
	}
}
