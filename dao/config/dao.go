package config

import (
	"github.com/cntechpower/utils/log"

	"github.com/cntechpower/anywhere/dao"
	"github.com/cntechpower/anywhere/model"
)

func Save(config *model.ProxyConfig) (err error) {
	err = dao.PersistDB().Save(config).Error
	return
}

func Remove(id uint) (err error) {
	err = dao.PersistDB().Delete(&model.ProxyConfig{}, "id = ?", id).Error
	return
}

// RemoveByKey deletes the row matching the natural key of a proxy config.
func RemoveByKey(userName, zoneName string, remotePort int, localAddr string) (err error) {
	err = dao.PersistDB().Where("user_name = ?", userName).
		Where("zone_name = ?", zoneName).
		Where("remote_port = ?", remotePort).
		Where("local_addr = ?", localAddr).
		Delete(&model.ProxyConfig{}).Error
	return
}

// Update only touches the whitelist columns. It must not use Save with a
// partially filled struct: Save would insert a new row instead of updating.
func Update(config *model.ProxyConfig) (err error) {
	err = dao.PersistDB().Model(&model.ProxyConfig{}).
		Where("user_name = ?", config.UserName).
		Where("zone_name = ?", config.ZoneName).
		Where("remote_port = ?", config.RemotePort).
		Where("local_addr = ?", config.LocalAddr).
		Updates(map[string]interface{}{
			"is_white_list_on": config.IsWhiteListOn,
			"white_cidr_list":  config.WhiteCidrList,
		}).Error
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
	h := log.NewHeader("MigrateFileToDB")
	cs, err := parseProxyConfigFile()
	if err != nil {
		return
	}
	for _, u := range cs.ProxyConfigs {
		for _, c := range u {
			if err = Save(c); err != nil {
				h.Errorf("save %+v to db error: %+v", c, err)
			}
		}
	}
	return
}

// IteratorByZone streams the proxy configs of a single zone instead of
// scanning the whole table for every zone.
func IteratorByZone(userName, zoneName string, fn func(c *model.ProxyConfig)) (err error) {
	res := make([]*model.ProxyConfig, 0)
	err = dao.PersistDB().Where("user_name = ?", userName).
		Where("zone_name = ?", zoneName).Find(&res).Error
	if err != nil {
		return err
	}
	for _, c := range res {
		fn(c)
	}
	return nil
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
