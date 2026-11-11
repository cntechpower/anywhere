package conf

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/pquerna/otp/totp"

	"github.com/cntechpower/utils/log"

	"github.com/cntechpower/anywhere/constants"
	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/util"
)

var Conf *model.SystemConfig

var (
	initConfig = &model.SystemConfig{
		ServerId: "anywhered-1",
		MainPort: 1111,
		AgentSsl: &model.SslConfig{
			CertFile: "credential/server.crt",
			KeyFile:  "credential/server.key",
			CaFile:   "credential/ca.crt",
		},
		HttpSSL: &model.SslConfig{
			CertFile: "credential/http.crt",
			KeyFile:  "credential/http.key",
			CaFile:   "credential/ca.crt",
		},
		UiConfig: &model.UiConfig{
			SkipLogin:   false,
			GrpcAddr:    "127.0.0.1:1113",
			IsWebEnable: true,
			WebAddr:     "0.0.0.0:1114",
			EnableTLS:   true,
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
	}
)

// Init loads and validates anywhered.json. Commands that need the config call
// it lazily so that `anywhered config reset` works on a fresh install.
func Init() error {
	gc, err := parseSystemConfigFile()
	if err != nil {
		return err
	}
	normalize(gc)
	Conf = gc
	return nil
}

// normalize fills in values that can be derived at runtime and warns about
// insecure configuration.
func normalize(c *model.SystemConfig) {
	if c.UiConfig.JwtSecret == "" {
		c.UiConfig.JwtSecret = randomSecret()
		log.Warnf(log.NewHeader("conf"), "ui_config.jwt_secret is empty, generated an ephemeral one; "+
			"UI sessions will be invalidated on restart. Add jwt_secret to anywhered.json for stable sessions.")
	}
	for _, u := range c.User.Users {
		if u == nil {
			continue
		}
		if u.UserPassHash == "" && u.UserPass == "admin" {
			log.Warnf(log.NewHeader("conf"), "user %q still uses the default password, change it before exposing the server", u.UserName)
		}
	}
}

func randomSecret() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("cannot generate jwt secret: " + err.Error())
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// JwtSecret returns the key used to sign UI session tokens.
func JwtSecret() []byte {
	if Conf == nil || Conf.UiConfig == nil || Conf.UiConfig.JwtSecret == "" {
		panic("jwt secret is not initialized")
	}
	return []byte(Conf.UiConfig.JwtSecret)
}

// SessionKey derives the cookie store key from the jwt secret so both rotate
// together and only one secret needs to be configured.
func SessionKey() []byte {
	sum := sha256.Sum256(append([]byte("anywhere-session:"), JwtSecret()...))
	return sum[:]
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
	return fmt.Errorf("%s is illegal in config section [%s], reason: %+v", printName, printSection, err)
}

func parseSystemConfigFile() (*model.SystemConfig, error) {
	file, err := os.Open(constants.SystemConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%s not found, run `anywhered config reset` to generate it", constants.SystemConfigFile)
		}
		return nil, err
	}
	defer func() { _ = file.Close() }()
	content, err := os.ReadFile(file.Name())
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
	if config.ProxyConnIdleTimeoutSeconds < 0 {
		return nil, newConfigIllegalError("main", "ProxyConnIdleTimeoutSeconds", fmt.Errorf("must be >= 0"))
	}
	if config.UiConfig == nil {
		return nil, newConfigMissedError("main", "UiConfig")
	}
	if config.UiConfig.GrpcAddr == "" {
		return nil, newConfigMissedError("UiConfig", "GrpcAddr")
	}
	if config.UiConfig.IsWebEnable {
		if err := util.CheckAddrValid(config.UiConfig.WebAddr); err != nil {
			return nil, newConfigIllegalError("UiConfig", "WebAddr", err)
		}
	}
	if config.User == nil || len(config.User.Users) == 0 {
		return nil, newConfigMissedError("main", "User")
	}
	for _, u := range config.User.Users {
		if u == nil || u.UserName == "" {
			return nil, newConfigMissedError("UserConfig", "UserName")
		}
		if u.UserPass == "" && u.UserPassHash == "" {
			return nil, newConfigMissedError("UserConfig", "UserPass")
		}
		if u.OtpEnable && u.OtpCode == "" {
			return nil, newConfigMissedError("UserConfig", "OtpCode")
		}
	}

	if config.AgentSsl == nil {
		return nil, newConfigMissedError("main", "AgentSsl")
	}
	if config.UiConfig.IsWebEnable && config.UiConfig.EnableTLS {
		if config.HttpSSL == nil {
			return nil, newConfigMissedError("main", "HttpSSL")
		}
		if config.HttpSSL.CertFile == "" {
			return nil, newConfigMissedError("HttpSSL", "CertFile")
		}
		if config.HttpSSL.KeyFile == "" {
			return nil, newConfigMissedError("HttpSSL", "KeyFile")
		}
		if !util.CheckPathExist(config.HttpSSL.CertFile) {
			return nil, newConfigIllegalError("HttpSSL", "CertFile", fmt.Errorf("file not exist"))
		}
		if !util.CheckPathExist(config.HttpSSL.KeyFile) {
			return nil, newConfigIllegalError("HttpSSL", "KeyFile", fmt.Errorf("file not exist"))
		}
	}
	if !util.CheckPathExist(config.AgentSsl.CaFile) {
		return nil, newConfigIllegalError("AgentSsl", "CaFile", fmt.Errorf("file not exist"))
	}
	if !util.CheckPathExist(config.AgentSsl.KeyFile) {
		return nil, newConfigIllegalError("AgentSsl", "KeyFile", fmt.Errorf("file not exist"))
	}
	if !util.CheckPathExist(config.AgentSsl.CertFile) {
		return nil, newConfigIllegalError("AgentSsl", "CertFile", fmt.Errorf("file not exist"))
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
	file, err := os.Create(constants.SystemConfigFile)
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

// WriteInitConfigFile writes a fresh config with random credentials. The
// generated password and OTP secret are printed once; they are not stored
// anywhere else.
func WriteInitConfigFile() error {
	cfg := *initConfig
	ui := *initConfig.UiConfig
	ui.JwtSecret = randomSecret()
	cfg.UiConfig = &ui

	password, err := randomPassword()
	if err != nil {
		return err
	}
	otpKey, err := totp.Generate(totp.GenerateOpts{Issuer: "anywhere", AccountName: "admin"})
	if err != nil {
		return err
	}
	cfg.User = &model.UserConfig{Users: []*model.User{{
		UserName:  "admin",
		UserPass:  password,
		IsAdmin:   true,
		OtpEnable: true,
		OtpCode:   otpKey.Secret(),
	}}}

	if err := WriteSystemConfigFile(&cfg); err != nil {
		return err
	}
	fmt.Printf("\nInitial admin credentials (store them now, they are not persisted anywhere else):\n"+
		"  username:   admin\n  password:   %s\n  otp secret: %s\n  otp url:    %s\n\n",
		password, otpKey.Secret(), otpKey.URL())
	return nil
}

func randomPassword() (string, error) {
	b := make([]byte, 18)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
