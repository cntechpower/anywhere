package config

import (
	"github.com/cntechpower/anywhere/dao"
	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/utils/log"
)

func Add(config *model.ProxyConfig) (err error) {
	err = dao.ConfigDB().Save(config).Error
	return
}

func Remove(id uint) (err error) {
	err = dao.ConfigDB().Delete(&model.ProxyConfig{}, "id = ?", id).Error
	return
}

func Update(userName, zoneName string, remotePort int, localAddr, whiteCidrs string, whiteListEnable bool) (err error) {
	err = dao.ConfigDB().Where("user_name=?", userName).
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
	err = dao.ConfigDB().Where("user_name=?", userName).
		Where("zone_name=?", zoneName).
		Where("remote_port=?", remotePort).Count(&count).Error
	if err != nil {
		return
	}
	exist = count >= 1
	return
}

func Migrate() (err error) {
	h := log.NewHeader("MigrateFileToDB")
	cs, err := parseProxyConfigFile()
	if err != nil {
		return
	}
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
	err = dao.ConfigDB().Find(&res).Error
	if err != nil {
		return err
	}
	for _, c := range res {
		tmpC := c
		fn(tmpC)
	}
	return
}

func GetById(id int64) (res *model.ProxyConfig, err error) {
	res = &model.ProxyConfig{}
	err = dao.ConfigDB().First(res, "id = ?", id).Error
	return
}