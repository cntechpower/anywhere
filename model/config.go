package model

import "gorm.io/gorm"

type SystemConfig struct {
	ServerId string      `json:"server_id"`
	MainPort int         `json:"server_port"`
	HttpSSL  *SslConfig  `json:"http_ssl_config"`
	AgentSsl *SslConfig  `json:"ssl_config"`
	UiConfig *UiConfig   `json:"ui_config"`
	User     *UserConfig `json:"user_config"`
	// ProxyConnIdleTimeoutSeconds closes proxied TCP connections that stay
	// completely idle for this long. 0 (default) disables the timeout, which is
	// what long lived tunnels such as SSH usually need.
	ProxyConnIdleTimeoutSeconds int `json:"proxy_conn_idle_timeout_seconds,omitempty"`
}

type UserConfig struct {
	Users []*User `json:"users"`
}

type User struct {
	gorm.Model `json:"-"`
	UserName   string `json:"user_name"`
	// UserPass is a legacy plain text password. Prefer UserPassHash.
	UserPass     string `json:"user_password,omitempty"`
	UserPassHash string `json:"user_password_hash,omitempty"`
	IsAdmin      bool   `json:"is_admin"`
	OtpEnable    bool   `json:"otp_enable"`
	OtpCode      string `json:"otp_code"`
}

type UiConfig struct {
	SkipLogin bool `json:"skip_login"`
	// JwtSecret signs UI session tokens. When empty a random secret is
	// generated at startup, which invalidates sessions on every restart.
	JwtSecret string `json:"jwt_secret,omitempty"`
	GrpcAddr  string `json:"grpc_addr"`
	// GrpcInsecure disables TLS on the local management gRPC endpoint. Keep it
	// false unless you are talking to an old server that only speaks plaintext.
	GrpcInsecure bool   `json:"grpc_insecure,omitempty"`
	IsWebEnable  bool   `json:"is_web_enable"`
	WebAddr      string `json:"web_ui_listen_addr"`
	DebugMode    bool   `json:"debug"`
	EnableTLS    bool   `json:"enable_tls"`
}

type SslConfig struct {
	CertFile string `json:"cert_file_path"`
	KeyFile  string `json:"key_file_path"`
	CaFile   string `json:"ca_file_path"`
}
