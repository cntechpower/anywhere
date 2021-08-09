package whitelist

import (
	"github.com/cntechpower/anywhere/model"

	"github.com/cntechpower/anywhere/dao"

	_ "github.com/go-sql-driver/mysql"
)

func AddWhiteListDenyIp(remotePort int, userName, zoneName, localAddr, ip string) (err error) {
	r := &model.WhiteListDenyRecord{
		UserName:   userName,
		ZoneName:   zoneName,
		RemotePort: remotePort,
		LocalAddr:  localAddr,
		IP:         ip,
	}
	err = dao.PersistDB().Save(r).Error
	return
}

func GetWhiteListDenyRank(limit int) (details []*model.WhiteListDenyRecord, err error) {
	details = make([]*model.WhiteListDenyRecord, 0)
	err = dao.PersistDB().Order("id DESC").Find(&details).Limit(limit).Error
	return
}
