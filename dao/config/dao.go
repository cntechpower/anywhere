package config

import (
	"github.com/cntechpower/anywhere/dao"
	"github.com/cntechpower/anywhere/model"
	log "github.com/cntechpower/utils/log.v2"
)

func Save(config *model.ProxyConfig) (err error) {
	err = dao.PersistDB().Save(config).Error
	return
}

func Remove(id uint) (err error) {
	err = dao.PersistDB().Delete(&model.ProxyConfig{}, "id = ?", id).Error
	return
}

func Update(config *model.ProxyConfig) (err error) {
	err = dao.PersistDB().Where("user_name=?", config.UserName).
		Where("zone_name=?", config.ZoneName).
		Where("remote_port=?", config.RemotePort).Save(config).Error
	return
}

func IsExist(userName, zoneName string, remotePort int) (exist bool, err error) {
	count := int64(0)
	err = dao.PersistDB().Where("user_name=?", userName).
		Where("zone_name=?", zoneName).
		Where("remote_port=?", remotePort).Count(&count).Error
	if err != nil {
		return
	}
	exist = count >= 1
	return
}

func Migrate() (err error) {
	fields := map[string]interface{}{
		log.FieldNameBizName: "dao.config.Migrate",
	}
	cs, err := parseProxyConfigFile()
	if err != nil {
		return
	}
	for _, u := range cs.ProxyConfigs {
		for _, c := range u {
			if err = Save(c); err != nil {
				log.Errorf(fields, "save %+v to db error: %v", c, err)
			}
		}
	}
	return
}

func Iterator(fn func(c *model.ProxyConfig)) (err error) {
	res := make([]*model.ProxyConfig, 0)
	err = dao.PersistDB().Find(&res).Error
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
	err = dao.PersistDB().First(res, "id = ?", id).Error
	return
}
