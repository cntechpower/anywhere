package model

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
