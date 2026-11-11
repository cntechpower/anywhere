package connlist

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cntechpower/anywhere/conn"
	"github.com/cntechpower/anywhere/internal/trace"
	"github.com/cntechpower/anywhere/model"
)

// connID is process wide so an ID returned by one zone never collides with an
// ID of another zone. The server uses that to kill a connection by ID without
// knowing its zone.
var connID atomic.Uint64

type joinedConn struct {
	item *model.JoinedConnListItem
	src  *conn.WrappedConn
	dst  *conn.WrappedConn
}

// JoinedConnList tracks the proxied connections of one (user, zone) pair.
// It is intentionally in-memory only: these records are ephemeral and writing
// two rows to SQLite per connection was a significant bottleneck at scale.
type JoinedConnList struct {
	userName string
	zoneName string
	list     map[uint]*joinedConn
	listMu   sync.RWMutex
}

func NewJoinedConnList(userName, zoneName string) *JoinedConnList {
	return &JoinedConnList{
		userName: userName,
		zoneName: zoneName,
		list:     make(map[uint]*joinedConn),
	}
}

func (l *JoinedConnList) Add(ctx context.Context, src, dst *conn.WrappedConn) uint {
	if src == nil || dst == nil {
		return 0
	}
	if ctx != nil {
		span, _ := trace.New(ctx, "JoinedConnList.Add")
		defer span.Finish()
	}

	item := &model.JoinedConnListItem{
		ID:            uint(connID.Add(1)),
		CreatedAt:     time.Now(),
		UserName:      l.userName,
		ZoneName:      l.zoneName,
		SrcName:       src.RemoteName,
		DstName:       dst.RemoteName,
		SrcRemoteAddr: src.GetRemoteAddr(),
		SrcLocalAddr:  src.GetLocalAddr(),
		DstRemoteAddr: dst.GetRemoteAddr(),
		DstLocalAddr:  dst.GetLocalAddr(),
	}

	l.listMu.Lock()
	l.list[item.ID] = &joinedConn{item: item, src: src, dst: dst}
	l.listMu.Unlock()
	return item.ID
}

// KillById closes the connections and removes the entry.
func (l *JoinedConnList) KillById(id uint) error {
	l.listMu.Lock()
	c, exist := l.list[id]
	if !exist {
		l.listMu.Unlock()
		return fmt.Errorf("no such id %d", id)
	}
	delete(l.list, id)
	l.listMu.Unlock()

	_ = c.src.Close()
	_ = c.dst.Close()
	return nil
}

// Remove forgets an entry without closing its connections (the caller owns
// the connection lifecycle, e.g. after JoinConn already closed them).
func (l *JoinedConnList) Remove(id uint) error {
	l.listMu.Lock()
	defer l.listMu.Unlock()
	if _, exist := l.list[id]; !exist {
		return fmt.Errorf("no such id %d", id)
	}
	delete(l.list, id)
	return nil
}

// Flush closes and forgets every connection of this list.
func (l *JoinedConnList) Flush() {
	l.listMu.Lock()
	conns := make([]*joinedConn, 0, len(l.list))
	for id, c := range l.list {
		conns = append(conns, c)
		delete(l.list, id)
	}
	l.listMu.Unlock()

	for _, c := range conns {
		_ = c.src.Close()
		_ = c.dst.Close()
	}
}

// List returns a snapshot ordered by ID (newest first).
func (l *JoinedConnList) List() []*model.JoinedConnListItem {
	l.listMu.RLock()
	res := make([]*model.JoinedConnListItem, 0, len(l.list))
	for _, c := range l.list {
		item := *c.item
		res = append(res, &item)
	}
	l.listMu.RUnlock()

	sort.Slice(res, func(i, j int) bool { return res[i].ID > res[j].ID })
	return res
}

// Count returns the number of connections of this list.
func (l *JoinedConnList) Count() int {
	l.listMu.RLock()
	defer l.listMu.RUnlock()
	return len(l.list)
}
