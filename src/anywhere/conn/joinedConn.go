package conn

import (
	"anywhere/util"
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
	list   map[int]*joinedConn
	listMu sync.RWMutex
}

func NewJoinedConnList() *JoinedConnList {
	return &JoinedConnList{
		list: make(map[int]*joinedConn, 0),
	}
}

func (l *JoinedConnList) Add(src, dst *BaseConn) int {
	l.listMu.Lock()
	defer l.listMu.Unlock()
	if l.list == nil {
		l.list = make(map[int]*joinedConn, 0)
	}
	idx := util.RandInt(9999)
	for {
		if c, exist := l.list[idx]; exist && c != nil {
			idx = util.RandInt(9999)
		} else {
			break
		}
	}

	l.list[idx] = &joinedConn{
		src: src,
		dst: dst,
	}
	return idx //return index

}

func (l *JoinedConnList) KillById(id int) error {
	if id < 0 {
		return fmt.Errorf("illegal id %v", id)
	}
	l.listMu.Lock()
	defer l.listMu.Unlock()
	if c, exist := l.list[id]; !exist {
		return fmt.Errorf("no such id %v", id)
	} else {
		c.src.Close()
		c.dst.Close()
	}
	return nil
}

func (l *JoinedConnList) Remove(id int) error {
	if id < 0 {
		return fmt.Errorf("illegal id %v", id)
	}
	l.listMu.Lock()
	defer l.listMu.Unlock()
	if _, exist := l.list[id]; !exist {
		return fmt.Errorf("no such id %v", id)
	} else {
		delete(l.list, id)
	}
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