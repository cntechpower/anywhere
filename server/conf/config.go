package conf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/cntechpower/anywhere/constants"
	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/util"
	"github.com/cntechpower/utils/log"
)

var proxyConf *ProxyConfigs
var Conf *model.SystemConfig
var configMu sync.RWMutex

type ProxyConfigs struct {
	ProxyConfigs map[string] /*user*/ []*model.ProxyConfig
}

func Add(config *model.ProxyConfig) error {
	if proxyConf == nil {
		return fmt.Errorf("config not init")
	}
	return proxyConf.Add(config)
}

func Remove(userName, zoneName string, remotePort int) error {
	if proxyConf == nil {
		return fmt.Errorf("config not init")
	}
	return proxyConf.Remove(userName, zoneName, remotePort)
}

func Update(userName, zoneName string, remotePort int, localAddr, whiteCidrs string, whiteListEnable bool) error {
	if proxyConf == nil {
		return fmt.Errorf("config not init")
	}
	return proxyConf.Update(userName, zoneName, remotePort, localAddr, whiteCidrs, whiteListEnable)
}

func PersistGlobalConfigLoop() {
	h := log.NewHeader("persist_config_loop")
	for range time.NewTicker(5 * time.Second).C {
		configMu.RLock()
		if proxyConf == nil {
			configMu.RUnlock()
			log.Infof(h, "skip because config is nil")
			continue
		}
		file, err := os.Create(constants.ProxyConfigFileName)
		if err != nil {
			configMu.RUnlock()
			log.Errorf(h, "create file %v error: %v", constants.ProxyConfigFileName, err)
			continue
		}
		bs, err := json.MarshalIndent(proxyConf, "", "    ")
		if err != nil {
			configMu.RUnlock()
			_ = file.Close()
			log.Errorf(h, "marshal config error: %v", err)
			continue
		}
		_, err = file.Write(bs)
		if err != nil {
			configMu.RUnlock()
			_ = file.Close()
			log.Errorf(h, "write config to file error: %v", err)
			continue
		}
		configMu.RUnlock()
		_ = file.Close()
	}
}

func Init() {
	h := log.NewHeader("init_proxy_config")
	var err error
	proxyConf, err = ParseProxyConfigFile()
	if err != nil {
		log.Warnf(h, "read config file %v error: %v, will start with empty config",
			constants.ProxyConfigFileName, err)
		proxyConf = &ProxyConfigs{}
	}
	if proxyConf.ProxyConfigs == nil {
		proxyConf.ProxyConfigs = make(map[string][]*model.ProxyConfig, 0)
	}
	Conf, err = parseSystemConfigFile()
	if err != nil {
		panic(err)
	}
	go PersistGlobalConfigLoop()
}

func (c *ProxyConfigs) ProxyConfigIterator(fn func(userName string, config *model.ProxyConfig) error) error {
	configMu.RLock()
	defer configMu.RUnlock()
	for userName, configs := range c.ProxyConfigs {
		for _, config := range configs {
			if err := fn(userName, config); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *ProxyConfigs) IsConfigExist(userName, zoneName string, remotePort int) bool {
	configMu.RLock()
	defer configMu.RUnlock()
	if _, ok := c.ProxyConfigs[userName]; !ok {
		c.ProxyConfigs[userName] = make([]*model.ProxyConfig, 0)
		return false
	}
	for _, config := range c.ProxyConfigs[userName] {
		if config.ZoneName == zoneName &&
			config.RemotePort == remotePort {
			return true
		}
	}
	return false
}

func (c *ProxyConfigs) Add(config *model.ProxyConfig) error {
	if c.IsConfigExist(config.UserName, config.ZoneName, config.RemotePort) {
		return fmt.Errorf("config for user: %v, group: %v, remotePort: %v already exist",
			config.UserName, config.ZoneName, config.RemotePort)
	}
	if _, ok := c.ProxyConfigs[config.UserName]; !ok {
		c.ProxyConfigs[config.UserName] = make([]*model.ProxyConfig, 0)
	}
	c.ProxyConfigs[config.UserName] = append(c.ProxyConfigs[config.UserName], config)
	return nil
}

func (c *ProxyConfigs) Remove(userName, zoneName string, remotePort int) error {
	if !c.IsConfigExist(userName, zoneName, remotePort) {
		return fmt.Errorf("config for user: %v, zoneName: %v, remotePort: %v not exist",
			userName, zoneName, remotePort)
	}
	for idx, config := range c.ProxyConfigs[userName] {
		if config.ZoneName == zoneName && config.RemotePort == remotePort {
			c.ProxyConfigs[userName] = append(c.ProxyConfigs[userName][:idx], c.ProxyConfigs[userName][idx+1:]...)
			return nil
		}
	}
	return nil
}

func (c *ProxyConfigs) Update(userName, zoneName string, remotePort int, localAddr, whiteCidrs string, whiteListEnable bool) error {
	if !c.IsConfigExist(userName, zoneName, remotePort) {
		return fmt.Errorf("config for user: %v, zoneName: %v, remotePort: %v not exist",
			userName, zoneName, remotePort)
	}
	for idx, config := range c.ProxyConfigs[userName] {
		if config.ZoneName == zoneName && config.RemotePort == remotePort {
			c.ProxyConfigs[userName][idx].LocalAddr = localAddr
			c.ProxyConfigs[userName][idx].WhiteCidrList = whiteCidrs
			c.ProxyConfigs[userName][idx].IsWhiteListOn = whiteListEnable
			return nil
		}
	}
	return nil
}

var (
	initConfig = &model.SystemConfig{
		ServerId:         "anywhered-1",
		MysqlDSN:         "anywhere:anywhere@tcp(10.0.0.2:3306)/anywhere?charset=utf8mb4&parseTime=True&loc=Local&timeout=5s&readTimeout=5s",
		MainPort:         1111,
		ReportCron:       "20 9 * * *",
		ReportWhiteCidrs: "180.168.0.0/16,101.86.210.116/16,223.88.249.186/16,120.253.0.0/16",
		AgentSsl: &model.SslConfig{
			CertFile: "credential/server.crt",
			KeyFile:  "credential/server.key",
			CaFile:   "credential/ca.crt",
		},
		UiConfig: &model.UiConfig{
			SkipLogin:   true,
			GrpcAddr:    "127.0.0.1:1113",
			IsWebEnable: true,
			RestAddr:    "127.0.0.1:1112",
			WebAddr:     "0.0.0.0:1114",
		},
		User: &model.UserConfig{
			Users: []*model.User{{
				UserName:  "admin",
				UserPass:  "admin",
				IsAdmin:   true,
				OtpEnable: false,
				OtpCode:   "ZKQVBFY55NJGGWBV5F6CU5CEK2YAWIB6",
			}},
		},
		SmtpConfig: &model.SmtpConfig{
			Host:     "smtp.exmail.qq.com",
			Port:     465,
			UserName: "no_reply@cntechpower.com",
			Password: "APB0K77gamkkAaFc",
		},
	}
)

func ParseProxyConfigFile() (*ProxyConfigs, error) {
	file, err := os.Open(constants.ProxyConfigFileName)
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	config := &ProxyConfigs{}
	if err := json.Unmarshal(content, config); err != nil {
		return nil, err
	}
	return config, nil
}

func getConfigJsonTag(sectionName, configName string) (string, string) {
	printName := configName
	printSection := sectionName
	if sectionName == "main" {
		field, ok := reflect.TypeOf(&model.SystemConfig{}).Elem().FieldByName(configName)
		if ok {
			if jsonTag := field.Tag.Get("json"); jsonTag != "" {
				printName = jsonTag
			}
		}
	} else {
		section, ok := reflect.TypeOf(&model.SystemConfig{}).Elem().FieldByName(sectionName)
		if ok {
			if jsonTag := section.Tag.Get("json"); jsonTag != "" {
				printSection = jsonTag
			}
			if section.Type.Elem().Kind() == reflect.Struct {
				configField, ok := section.Type.Elem().FieldByName(configName)
				if ok {
					if jsonTag := configField.Tag.Get("json"); jsonTag != "" {
						printName = jsonTag
					}
				}
			}

		}

	}
	return printSection, printName
}

func newConfigMissedError(sectionName, configName string) error {
	printSection, printName := getConfigJsonTag(sectionName, configName)
	return fmt.Errorf("%s is required in config section [%s], use `./anywhered config reset` to get default config file", printName, printSection)
}

func newConfigIllegalError(sectionName, configName string, err error) error {
	printSection, printName := getConfigJsonTag(sectionName, configName)
	return fmt.Errorf("%s is illegal in config section [%s], reason: %v", printName, printSection, err)
}

func parseSystemConfigFile() (*model.SystemConfig, error) {
	file, err := os.Open(constants.SystemConfigFIle)
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	config := &model.SystemConfig{}
	if err := json.Unmarshal(content, config); err != nil {
		return nil, err
	}
	if config.ServerId == "" {
		return nil, newConfigMissedError("main", "ServerId")
	}
	if config.MainPort == 0 {
		return nil, newConfigMissedError("main", "MainPort")
	}
	if config.UiConfig != nil {
		if config.UiConfig.GrpcAddr == "" {
			return nil, newConfigMissedError("UiConfig", "GrpcAddr")
		}
		if err := util.CheckAddrValid(config.UiConfig.WebAddr); err != nil {
			return nil, newConfigIllegalError("UiConfig", "WebAddr", err)
		}
		if err := util.CheckAddrValid(config.UiConfig.RestAddr); err != nil {
			return nil, newConfigIllegalError("UiConfig", "RestAddr", err)
		}
		if config.User == nil {
			return nil, newConfigMissedError("main", "User")
		}
	}

	if config.AgentSsl != nil {
		if !util.CheckPathExist(config.AgentSsl.CaFile) {
			return nil, newConfigIllegalError("AgentSsl", "CaFile", fmt.Errorf("file not exist"))
		}
		if !util.CheckPathExist(config.AgentSsl.KeyFile) {
			return nil, newConfigIllegalError("AgentSsl", "KeyFile", fmt.Errorf("file not exist"))
		}
		if !util.CheckPathExist(config.AgentSsl.CertFile) {
			return nil, newConfigIllegalError("AgentSsl", "CertFile", fmt.Errorf("file not exist"))
		}
	}

	return config, nil
}

func GetGrpcAddr() (string, error) {
	c, err := parseSystemConfigFile()
	if err != nil {
		return "", err
	}
	return c.UiConfig.GrpcAddr, nil
}

func WriteSystemConfigFile(config *model.SystemConfig) error {
	if config == nil {
		return nil
	}
	file, err := os.Create(constants.SystemConfigFIle)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	bs, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}
	_, err = file.Write(bs)
	return err
}

func WriteInitConfigFile() error {
	return WriteSystemConfigFile(initConfig)
}
