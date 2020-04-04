package conn

import (
	"fmt"
	"sync"
)

type joinedConn struct {
	src *BaseConn
	dst *BaseConn
}

type JoinedConnListItem struct {
	ConnId        int
	SrcRemoteAddr string
	SrcLocalAddr  string
	DstRemoteAddr string
	DstLocalAddr  string
}

type JoinedConnList struct {
	list   []*joinedConn
	listMu sync.RWMutex
}

func NewJoinedConnList() *JoinedConnList {
	return &JoinedConnList{
		list: make([]*joinedConn, 0),
	}
}

func (l *JoinedConnList) Add(src, dst *BaseConn) int {
	l.listMu.Lock()
	defer l.listMu.Unlock()
	if l.list == nil {
		l.list = make([]*joinedConn, 0)
	}
	l.list = append(l.list, &joinedConn{
		src: src,
		dst: dst,
	})
	return len(l.list) - 1 //return index

}

func (l *JoinedConnList) KillById(id int) error {
	if id < 0 {
		return fmt.Errorf("illegal id %v", id)
	}
	l.listMu.Lock()
	defer l.listMu.Unlock()
	if len(l.list) < (id + 1) {
		return fmt.Errorf("no such joinedConn")
	}
	l.list[id].src.Close()
	l.list[id].dst.Close()
	return nil
}

func (l *JoinedConnList) Remove(id int) error {
	if id < 0 {
		return fmt.Errorf("illegal id %v", id)
	}
	l.listMu.Lock()
	defer l.listMu.Unlock()
	if len(l.list) < (id + 1) {
		return fmt.Errorf("no such joinedConn %v", id)
	}
	l.list = append(l.list[:id], l.list[id+1:]...)
	return nil
}

func (l *JoinedConnList) Flush() {
	l.listMu.Lock()
	defer l.listMu.Unlock()
	for _, joinedConn := range l.list {
		joinedConn.src.Close()
		joinedConn.dst.Close()
	}
}

func (l *JoinedConnList) List() []*JoinedConnListItem {
	res := make([]*JoinedConnListItem, 0)
	for idx, conn := range l.list {
		res = append(res, &JoinedConnListItem{
			ConnId:        idx,
			SrcRemoteAddr: conn.src.GetRemoteAddr(),
			SrcLocalAddr:  conn.src.GetLocalAddr(),
			DstRemoteAddr: conn.dst.GetRemoteAddr(),
			DstLocalAddr:  conn.dst.GetLocalAddr(),
		})
	}
	return res
}
