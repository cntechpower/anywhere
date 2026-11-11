package conn

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

func TestFrameRoundTrip(t *testing.T) {
	cases := [][]byte{
		{},
		[]byte("hello"),
		bytes.Repeat([]byte("x"), 4096),
		bytes.Repeat([]byte("y"), MaxFrameSize),
	}
	for _, payload := range cases {
		var buf bytes.Buffer
		if err := writeFrame(&buf, payload); err != nil {
			t.Fatalf("writeFrame(%d bytes): %v", len(payload), err)
		}
		got, err := readFrame(&buf)
		if err != nil {
			t.Fatalf("readFrame(%d bytes): %v", len(payload), err)
		}
		if !bytes.Equal(got, payload) {
			t.Fatalf("payload mismatch: got %d bytes, want %d", len(got), len(payload))
		}
	}
}

func TestFrameRejectsOversizedPayload(t *testing.T) {
	payload := make([]byte, MaxFrameSize+1)
	var buf bytes.Buffer
	if err := writeFrame(&buf, payload); err == nil {
		t.Fatal("expected writeFrame to reject oversized payload")
	}

	hdr := []byte{0xff, 0xff, 0xff, 0x7f} // 2GiB
	if _, err := readFrame(bytes.NewReader(hdr)); err == nil {
		t.Fatal("expected readFrame to reject oversized frame")
	}
}

// TestFrameHandlesPartialReads verifies the receiver no longer depends on the
// whole frame being buffered already (the old implementation returned
// "size not enough" for messages split across TCP segments).
func TestFrameHandlesPartialReads(t *testing.T) {
	var encoded bytes.Buffer
	payload := bytes.Repeat([]byte("z"), 10000)
	if err := writeFrame(&encoded, payload); err != nil {
		t.Fatal(err)
	}
	full := encoded.Bytes()

	pr, pw := io.Pipe()
	go func() {
		for i := 0; i < len(full); i += 97 {
			end := i + 97
			if end > len(full) {
				end = len(full)
			}
			if _, err := pw.Write(full[i:end]); err != nil {
				return
			}
			time.Sleep(time.Millisecond)
		}
		_ = pw.Close()
	}()

	got, err := readFrame(pr)
	if err != nil {
		t.Fatalf("readFrame on segmented stream: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatal("payload mismatch")
	}
}

// TestReceiveDoesNotLoseBackToBackMessages guards against the old bug where a
// fresh bufio.Reader was created per Receive call, dropping prefetched bytes.
func TestReceiveDoesNotLoseBackToBackMessages(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	wc := NewWrappedConn("test", server)
	go func() {
		for i := 0; i < 3; i++ {
			_ = writeFrame(client, []byte(`{"n":`+string(rune('0'+i))+`}`))
		}
	}()

	for i := 0; i < 3; i++ {
		var msg map[string]int
		if err := wc.Receive(&msg); err != nil {
			t.Fatalf("receive #%d: %v", i, err)
		}
		if msg["n"] != i {
			t.Fatalf("receive #%d got %d", i, msg["n"])
		}
	}
}

func TestConcurrentSendDoesNotInterleave(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	wc := NewWrappedConn("test", server)
	const n = 50

	seen := make(chan int, n)
	go func() {
		for i := 0; i < n; i++ {
			payload, err := readFrame(client)
			if err != nil {
				close(seen)
				return
			}
			var msg map[string]int
			if err := json.Unmarshal(payload, &msg); err != nil {
				close(seen)
				return
			}
			seen <- msg["i"]
		}
		close(seen)
	}()

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if err := wc.Send(map[string]int{"i": i}); err != nil {
				t.Errorf("send: %v", err)
			}
		}(i)
	}
	wg.Wait()

	count := 0
	for range seen {
		count++
	}
	if count != n {
		t.Fatalf("received %d frames, want %d", count, n)
	}
}

func TestCloseUnblocksReceive(t *testing.T) {
	server, client := net.Pipe()
	wc := NewWrappedConn("test", server)

	done := make(chan error, 1)
	go func() {
		var msg map[string]int
		done <- wc.Receive(&msg)
	}()

	time.Sleep(20 * time.Millisecond)
	_ = wc.Close()
	_ = client.Close()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected error after close")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Receive did not unblock after Close")
	}
}

func TestResetConn(t *testing.T) {
	a, b := net.Pipe()
	wc := NewWrappedConn("test", a)
	go func() { _ = b.Close() }()
	time.Sleep(10 * time.Millisecond)

	c, d := net.Pipe()
	defer c.Close()
	defer d.Close()
	wc.ResetConn(c)
	if !wc.IsValid() {
		t.Fatal("expected valid conn after ResetConn")
	}

	go func() {
		_ = writeFrame(d, []byte(`{"ok":true}`))
	}()
	var msg map[string]bool
	if err := wc.Receive(&msg); err != nil {
		t.Fatalf("receive after reset: %v", err)
	}
	if !msg["ok"] {
		t.Fatal("unexpected payload")
	}
}

func TestJoinConnIdleTimeout(t *testing.T) {
	a1, a2 := net.Pipe()
	b1, b2 := net.Pipe()
	defer a2.Close()
	defer b2.Close()

	done := make(chan struct{})
	go func() {
		JoinConnWithIdleTimeout(a1, b1, 50*time.Millisecond)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("JoinConnWithIdleTimeout did not enforce idle timeout")
	}
}
