package connlist

import (
	"context"
	"net"
	"testing"

	"github.com/cntechpower/anywhere/conn"
)

func newTestConn(t *testing.T) (*conn.WrappedConn, net.Conn) {
	t.Helper()
	a, b := net.Pipe()
	return conn.NewWrappedConn("test", a), b
}

func TestJoinedConnListAddRemove(t *testing.T) {
	l := NewJoinedConnList("user", "zone")
	src, srcPeer := newTestConn(t)
	dst, dstPeer := newTestConn(t)
	defer srcPeer.Close()
	defer dstPeer.Close()

	id := l.Add(context.Background(), src, dst)
	if id == 0 {
		t.Fatal("expected non zero id")
	}
	if l.Count() != 1 {
		t.Fatalf("count = %d, want 1", l.Count())
	}
	list := l.List()
	if len(list) != 1 || list[0].ID != id || list[0].UserName != "user" || list[0].ZoneName != "zone" {
		t.Fatalf("unexpected list: %+v", list)
	}

	if err := l.Remove(id); err != nil {
		t.Fatal(err)
	}
	if l.Count() != 0 {
		t.Fatalf("count = %d, want 0", l.Count())
	}
	if err := l.Remove(id); err == nil {
		t.Fatal("expected error removing a missing id")
	}
}

func TestJoinedConnListIDsAreUniqueAcrossLists(t *testing.T) {
	l1 := NewJoinedConnList("u1", "z1")
	l2 := NewJoinedConnList("u2", "z2")

	src1, p1 := newTestConn(t)
	dst1, p2 := newTestConn(t)
	defer p1.Close()
	defer p2.Close()
	src2, p3 := newTestConn(t)
	dst2, p4 := newTestConn(t)
	defer p3.Close()
	defer p4.Close()

	id1 := l1.Add(nil, src1, dst1)
	id2 := l2.Add(nil, src2, dst2)
	if id1 == id2 {
		t.Fatalf("expected unique ids, got %d twice", id1)
	}
}

func TestJoinedConnListCountIsPerList(t *testing.T) {
	l1 := NewJoinedConnList("u1", "z1")
	l2 := NewJoinedConnList("u2", "z2")

	src, p1 := newTestConn(t)
	dst, p2 := newTestConn(t)
	defer p1.Close()
	defer p2.Close()
	l1.Add(nil, src, dst)

	if got := l2.Count(); got != 0 {
		t.Fatalf("other list count = %d, want 0", got)
	}
}

func TestJoinedConnListKillClosesBothSides(t *testing.T) {
	l := NewJoinedConnList("user", "zone")
	src, srcPeer := newTestConn(t)
	dst, dstPeer := newTestConn(t)
	defer srcPeer.Close()
	defer dstPeer.Close()

	id := l.Add(nil, src, dst)
	if err := l.KillById(id); err != nil {
		t.Fatal(err)
	}
	if src.IsValid() || dst.IsValid() {
		t.Fatal("expected both connections to be closed")
	}
}

func TestJoinedConnListFlushOnlyAffectsItself(t *testing.T) {
	l1 := NewJoinedConnList("u1", "z1")
	l2 := NewJoinedConnList("u2", "z2")

	src1, p1 := newTestConn(t)
	dst1, p2 := newTestConn(t)
	defer p1.Close()
	defer p2.Close()
	src2, p3 := newTestConn(t)
	dst2, p4 := newTestConn(t)
	defer p3.Close()
	defer p4.Close()

	l1.Add(nil, src1, dst1)
	l2.Add(nil, src2, dst2)

	l1.Flush()
	if l1.Count() != 0 {
		t.Fatalf("flushed list count = %d, want 0", l1.Count())
	}
	if l2.Count() != 1 {
		t.Fatalf("other list count = %d, want 1", l2.Count())
	}
	if !src2.IsValid() {
		t.Fatal("flush must not close connections of another list")
	}
}
