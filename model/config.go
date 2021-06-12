package model

type SystemConfig struct {
	ServerId         string      `json:"server_id"`
	MainPort         int         `json:"server_port"`
	ReportCron       string      `json:"report_cron"`
	ReportWhiteCidrs string      `json:"report_white_cidrs"`
	MysqlDSN         string      `json:"mysql_dsn"`
	HttpSSL          *SslConfig  `json:"http_ssl_config"`
	AgentSsl         *SslConfig  `json:"ssl_config"`
	UiConfig         *UiConfig   `json:"ui_config"`
	User             *UserConfig `json:"user_config"`
	SmtpConfig       *SmtpConfig `json:"smtp_config"`
}

type UserConfig struct {
	Users []*User `json:"users"`
}

type User struct {
	UserName  string `json:"user_name"`
	UserPass  string `json:"user_password"`
	IsAdmin   bool   `json:"is_admin"`
	OtpEnable bool   `json:"otp_enable"`
	OtpCode   string `json:"otp_code"`
}

type UiConfig struct {
	SkipLogin   bool   `json:"skip_login"`
	GrpcAddr    string `json:"grpc_addr"`
	IsWebEnable bool   `json:"is_web_enable"`
	RestAddr    string `json:"rest_api_listen_addr"`
	WebAddr     string `json:"web_ui_listen_addr"`
	DebugMode   bool   `json:"debug"`
}

type SslConfig struct {
	CertFile string `json:"cert_file_path"`
	KeyFile  string `json:"key_file_path"`
	CaFile   string `json:"ca_file_path"`
}

type SmtpConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	UserName string `json:"user_name"`
	Password string `json:"password"`
}
