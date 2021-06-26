package conn

import (
	"fmt"
	"sync"

	"github.com/cntechpower/utils/log"

	"github.com/cntechpower/anywhere/server/db"

	"github.com/cntechpower/anywhere/model"
)

type joinedConn struct {
	src *WrappedConn
	dst *WrappedConn
}

type JoinedConnList struct {
	name   string
	list   map[uint]*joinedConn
	listMu sync.RWMutex
}

func NewJoinedConnList(name string) *JoinedConnList {
	return &JoinedConnList{
		name: name,
		list: make(map[uint]*joinedConn, 0),
	}
}

func (l *JoinedConnList) Add(src, dst *WrappedConn) uint {
	l.listMu.Lock()
	defer l.listMu.Unlock()
	if l.list == nil {
		l.list = make(map[uint]*joinedConn, 0)
	}
	item := &model.JoinedConnListItem{
		Name:          l.name,
		SrcName:       src.remoteName,
		DstName:       dst.remoteName,
		SrcRemoteAddr: src.conn.RemoteAddr().String(),
		SrcLocalAddr:  src.conn.LocalAddr().String(),
		DstRemoteAddr: dst.conn.RemoteAddr().String(),
		DstLocalAddr:  dst.conn.LocalAddr().String(),
	}
	err := db.MemDB.Save(&item)

	if err != nil {
		log.NewHeader("JoinedConnList").Errorf("add error: %v", err)
	}

	l.list[item.ID] = &joinedConn{
		src: src,
		dst: dst,
	}
	return item.ID //return index

}

func (l *JoinedConnList) KillById(id uint) error {
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

func (l *JoinedConnList) Remove(id uint) error {
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
	db.MemDB.Delete(&model.JoinedConnListItem{}, "")
	for _, joinedConn := range l.list {
		joinedConn.src.Close()
		joinedConn.dst.Close()
	}
}

func (l *JoinedConnList) List() []*model.JoinedConnListItem {
	l.listMu.Lock()
	defer l.listMu.Unlock()
	res := make([]*model.JoinedConnListItem, 0)
	for idx, conn := range l.list {
		res = append(res, &model.JoinedConnListItem{
			ConnId:        idx,
			SrcName:       conn.src.remoteName,
			DstName:       conn.dst.remoteName,
			SrcRemoteAddr: conn.src.GetRemoteAddr(),
			SrcLocalAddr:  conn.src.GetLocalAddr(),
			DstRemoteAddr: conn.dst.GetRemoteAddr(),
			DstLocalAddr:  conn.dst.GetLocalAddr(),
		})
	}
	return res
}

func (l *JoinedConnList) Count() int {
	l.listMu.Lock()
	defer l.listMu.Unlock()
	return len(l.list)
}
