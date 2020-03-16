package anywhereServer

import (
	"anywhere/model"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

var initConfig = &model.SystemConfig{
	ServerId: "anywhered-1",
	Ssl: &model.SslConfig{
		CertFile: "credential/server.crt",
		KeyFile:  "credential/server.key",
		CaFile:   "credential/ca.crt",
	},
	Net: &model.NetworkConfig{
		MainPort:    1111,
		GrpcPort:    1113,
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

func (s *anyWhereServer) SaveConfigToFile() error {
	configs := s.ListProxyConfigs()
	if len(configs) == 0 {
		return fmt.Errorf("anywhere server is with no config, skip")
	}
	return writeProxyConfigFile(configs)
}

func (s *anyWhereServer) LoadProxyConfigFile() error {
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
	return config, nil
}

func GetGrpcPort() (int, error) {
	c, err := ParseSystemConfigFile()
	if err != nil {
		return 0, err
	}
	return c.Net.GrpcPort, nil
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
