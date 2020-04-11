package anywhereServer

import (
	"anywhere/model"
	"anywhere/util"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
)

var initConfig = &model.SystemConfig{
	ServerId: "anywhered-1",
	MainPort: 1111,
	Ssl: &model.SslConfig{
		CertFile: "credential/server.crt",
		KeyFile:  "credential/server.key",
		CaFile:   "credential/ca.crt",
	},
	UiConfig: &model.UiConfig{
		SkipLogin:   false,
		GrpcAddr:    "127.0.0.1:1113",
		IsWebEnable: true,
		RestAddr:    "127.0.0.1:1112",
		WebAddr:     "0.0.0.0:1114",
	},
	User: &model.UserConfig{
		AdminUser:      "admin",
		AdminPass:      "admin",
		AdminOtpEnable: false,
		AdminOtpCode:   "ZKQVBFY55NJGGWBV5F6CU5CEK2YAWIB6",
	},
}

func (s *Server) SaveConfigToFile() error {
	configs := s.ListProxyConfigs()
	if len(configs) == 0 {
		return fmt.Errorf("anywhere server is with no config, skip")
	}
	return writeProxyConfigFile(configs)
}

func (s *Server) LoadProxyConfigFile() error {
	configs, err := ParseProxyConfigFile()
	if err != nil {
		return err
	}
	for _, config := range configs.ProxyConfigs {
		if err := s.AddProxyConfigToAgentByModel(config); err != nil {
			return err
		}
	}
	return nil
}

const configFile = "proxy.json"
const systemConfigFIle = "anywhered.json"

func ParseProxyConfigFile() (*model.GlobalConfig, error) {
	file, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	config := &model.GlobalConfig{}
	if err := json.Unmarshal(content, config); err != nil {
		return nil, err
	}
	return config, nil
}

func writeProxyConfigFile(configs []*model.ProxyConfig) error {
	if configs == nil {
		return nil
	}
	file, err := os.Create(configFile)
	if err != nil {
		return err
	}
	globalConfig := &model.GlobalConfig{ProxyConfigs: make([]*model.ProxyConfig, 0)}
	for _, config := range configs {
		globalConfig.ProxyConfigs = append(globalConfig.ProxyConfigs, config)
	}
	bs, err := json.MarshalIndent(globalConfig, "", "    ")
	if err != nil {
		return err
	}
	_, err = file.Write(bs)
	return err

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

func ParseSystemConfigFile() (*model.SystemConfig, error) {
	file, err := os.Open(systemConfigFIle)
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

	if config.Ssl != nil {
		if !util.CheckPathExist(config.Ssl.CaFile) {
			return nil, newConfigIllegalError("Ssl", "CaFile", fmt.Errorf("file not exist"))
		}
		if !util.CheckPathExist(config.Ssl.KeyFile) {
			return nil, newConfigIllegalError("Ssl", "KeyFile", fmt.Errorf("file not exist"))
		}
		if !util.CheckPathExist(config.Ssl.CertFile) {
			return nil, newConfigIllegalError("Ssl", "CertFile", fmt.Errorf("file not exist"))
		}
	}

	return config, nil
}

func GetGrpcAddr() (string, error) {
	c, err := ParseSystemConfigFile()
	if err != nil {
		return "", err
	}
	return c.UiConfig.GrpcAddr, nil
}

func WriteSystemConfigFile(config *model.SystemConfig) error {
	if config == nil {
		return nil
	}
	file, err := os.Create(systemConfigFIle)
	if err != nil {
		return err
	}

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
