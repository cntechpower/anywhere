package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cntechpower/anywhere/constants"
	"github.com/cntechpower/anywhere/model"
)

func parseProxyConfigFile() (*model.ProxyConfigs, error) {
	content, err := os.ReadFile(constants.ProxyConfigFileName)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", constants.ProxyConfigFileName, err)
	}
	config := &model.ProxyConfigs{}
	if err := json.Unmarshal(content, config); err != nil {
		return nil, err
	}
	return config, nil
}
