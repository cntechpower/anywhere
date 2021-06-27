package conf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"

	"github.com/cntechpower/anywhere/constants"
	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/util"
)

var Conf *model.SystemConfig

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

func Init() {
	gc, err := parseSystemConfigFile()
	if err != nil {
		panic(err)
	}
	Conf = gc
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
