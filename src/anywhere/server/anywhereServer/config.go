package anywhereServer

import (
	"anywhere/model"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func (s *anyWhereServer) SaveConfigToFile() error {
	configs := s.ListProxyConfigs()
	if len(configs) == 0 {
		return fmt.Errorf("anywhere server is with no config, skip")
	}
	return writeConfigFile(configs)
}

func (s *anyWhereServer) LoadProxyConfigFile() error {
	configs, err := parseConfigFile()
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

const configFile = "anywhered.json"

func parseConfigFile() (*model.GlobalConfig, error) {
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

func writeConfigFile(configs []*model.ProxyConfig) error {
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
