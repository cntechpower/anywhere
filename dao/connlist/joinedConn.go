package connlist

import (
	"fmt"
	"sync"

	"github.com/cntechpower/anywhere/conn"

	"github.com/cntechpower/anywhere/dao"

	"github.com/cntechpower/utils/log"

	"github.com/cntechpower/anywhere/model"
)

type joinedConn struct {
	src *conn.WrappedConn
	dst *conn.WrappedConn
}

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
		list:     make(map[uint]*joinedConn, 0),
	}
}

func (l *JoinedConnList) Add(src, dst *conn.WrappedConn) uint {
	l.listMu.Lock()
	defer l.listMu.Unlock()
	if l.list == nil {
		l.list = make(map[uint]*joinedConn, 0)
	}
	item := &model.JoinedConnListItem{
		UserName:      l.userName,
		ZoneName:      l.zoneName,
		SrcName:       src.RemoteName,
		DstName:       dst.RemoteName,
		SrcRemoteAddr: src.Conn.RemoteAddr().String(),
		SrcLocalAddr:  src.Conn.LocalAddr().String(),
		DstRemoteAddr: dst.Conn.RemoteAddr().String(),
		DstLocalAddr:  dst.Conn.LocalAddr().String(),
	}
	err := dao.MemDB().Create(&item).Error

	if err != nil {
		log.NewHeader("JoinedConnList").Errorf("add error: %+v", err)
	}

	l.list[item.ID] = &joinedConn{
		src: src,
		dst: dst,
	}
	return item.ID //return index

}

func (l *JoinedConnList) KillById(id uint) (err error) {
	if id < 0 {
		return fmt.Errorf("illegal id %v", id)
	}
	l.listMu.Lock()
	defer l.listMu.Unlock()
	if c, exist := l.list[id]; !exist {
		return fmt.Errorf("no such id %v", id)
	} else {
		_ = c.src.Close()
		_ = c.dst.Close()
	}
	err = dao.MemDB().Delete(&model.JoinedConnListItem{}, "id = ?", id).Error
	return err
}

func (l *JoinedConnList) Remove(id uint) (err error) {
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
	err = dao.MemDB().Delete(&model.JoinedConnListItem{}, "id = ?", id).Error
	return nil
}

func (l *JoinedConnList) Flush() {
	l.listMu.Lock()
	defer l.listMu.Unlock()
	dao.MemDB().Delete(&model.JoinedConnListItem{}, "user_name = ?", l.userName).Where("zone_name = ?", l.zoneName)
	for _, joinedConn := range l.list {
		_ = joinedConn.src.Close()
		_ = joinedConn.dst.Close()
	}
}

func (l *JoinedConnList) List() (res []*model.JoinedConnListItem, err error) {
	l.listMu.Lock()
	defer l.listMu.Unlock()
	res = make([]*model.JoinedConnListItem, 0)
	err = dao.MemDB().Find(&res, "user_name = ?", l.userName).Where("zone_name = ?", l.zoneName).Error
	return
}

func (l *JoinedConnList) Count() (count int64, err error) {
	err = dao.MemDB().Model(&model.JoinedConnListItem{}).Count(&count).Error
	return
}

func GetJoinedConnList(userName, zoneName string) (res []*model.JoinedConnListItem, err error) {
	res = make([]*model.JoinedConnListItem, 0)
	db := dao.MemDB()
	if userName != "" {
		db = db.Where("user_name = ?", userName)
	}
	if zoneName != "" {
		db = db.Where("zone_name = ?", zoneName)
	}
	err = db.Find(&res).Error
	return
}

func GetJoinedConnById(id int64) (res *model.JoinedConnListItem, err error) {
	res = &model.JoinedConnListItem{}
	err = dao.MemDB().First(res, "id = ?", id).Error
	return
}
