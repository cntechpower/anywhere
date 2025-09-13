package model

import "gorm.io/gorm"

type SystemConfig struct {
	ServerId string      `json:"server_id"`
	MainPort int         `json:"server_port"`
	HttpSSL  *SslConfig  `json:"http_ssl_config"`
	AgentSsl *SslConfig  `json:"ssl_config"`
	UiConfig *UiConfig   `json:"ui_config"`
	User     *UserConfig `json:"user_config"`
}

type UserConfig struct {
	Users []*User `json:"users"`
}

type User struct {
	gorm.Model `json:"-"`
	UserName   string `json:"user_name"`
	UserPass   string `json:"user_password"`
	IsAdmin    bool   `json:"is_admin"`
	OtpEnable  bool   `json:"otp_enable"`
	OtpCode    string `json:"otp_code"`
}

type UiConfig struct {
	SkipLogin   bool   `json:"skip_login"`
	GrpcAddr    string `json:"grpc_addr"`
	IsWebEnable bool   `json:"is_web_enable"`
	WebAddr     string `json:"web_ui_listen_addr"`
	DebugMode   bool   `json:"debug"`
	EnableTLS   bool   `json:"enable_tls"`
}

type SslConfig struct {
	CertFile string `json:"cert_file_path"`
	KeyFile  string `json:"key_file_path"`
	CaFile   string `json:"ca_file_path"`
}
