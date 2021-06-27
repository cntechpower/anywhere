package config

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/cntechpower/anywhere/constants"
	"github.com/cntechpower/anywhere/model"
)

func parseProxyConfigFile() (*model.ProxyConfigs, error) {
	file, err := os.Open(constants.ProxyConfigFileName)
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	config := &model.ProxyConfigs{}
	if err := json.Unmarshal(content, config); err != nil {
		return nil, err
	}
	return config, nil
}
