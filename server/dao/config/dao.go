package config

import (
	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/server/db"
	"github.com/cntechpower/utils/log"
)

func Add(config *model.ProxyConfig) (err error) {
	err = db.ConfigDB.Save(config).Error
	return
}

func Remove(userName, zoneName string, remotePort int) (err error) {
	err = db.ConfigDB.Where("user_name=?", userName).
		Where("zone_name=?", zoneName).
		Where("remote_port=?", remotePort).Delete(&model.ProxyConfig{}).Error
	return
}

func Update(userName, zoneName string, remotePort int, localAddr, whiteCidrs string, whiteListEnable bool) (err error) {
	err = db.ConfigDB.Where("user_name=?", userName).
		Where("zone_name=?", zoneName).
		Where("remote_port=?", remotePort).Save(&model.ProxyConfig{
		UserName:      userName,
		ZoneName:      zoneName,
		RemotePort:    remotePort,
		LocalAddr:     localAddr,
		IsWhiteListOn: whiteListEnable,
		WhiteCidrList: whiteCidrs,
	}).Error
	return
}

func IsExist(userName, zoneName string, remotePort int) (exist bool, err error) {
	count := int64(0)
	err = db.ConfigDB.Where("user_name=?", userName).
		Where("zone_name=?", zoneName).
		Where("remote_port=?", remotePort).Count(&count).Error
	if err != nil {
		return
	}
	exist = count >= 1
	return
}

func Migrate(cs *model.ProxyConfigs) (err error) {
	h := log.NewHeader("MigrateFileToDB")
	for _, u := range cs.ProxyConfigs {
		for _, c := range u {
			if err = Add(c); err != nil {
				h.Errorf("save %+v to db error: %v", c, err)
			}
		}
	}
	return
}

func Iterator(fn func(c *model.ProxyConfig)) (err error) {
	res := make([]*model.ProxyConfig, 0)
	err = db.ConfigDB.Find(&res).Error
	if err != nil {
		return err
	}
	for _, c := range res {
		tmpC := c
		fn(tmpC)
	}
	return
}
