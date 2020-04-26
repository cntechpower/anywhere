package auth

import (
	"anywhere/log"

	"github.com/pquerna/otp/totp"
)

type TOTPValidator struct {
	otpSecretMap map[string] /*userName*/ string /*secret*/
	logHeader    *log.Header
	enable       bool
}

func NewTOTPValidator(adminUser, adminSecret string, enable bool) *TOTPValidator {
	return &TOTPValidator{
		otpSecretMap: map[string]string{
			adminUser: adminSecret,
		},
		logHeader: log.NewHeader("TOTPValidator"),
		enable:    enable,
	}
}

func (v *TOTPValidator) Generate(userName string) (string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "cntechpower.com",
		AccountName: userName,
	})
	if err != nil {
		log.Infof(v.logHeader, "generate totp secret for user %v fail", userName)
		return "", err
	}
	log.Infof(v.logHeader, "generate totp secret for user %v success", userName)
	v.otpSecretMap[userName] = key.Secret()
	return key.Secret(), nil
}

func (v *TOTPValidator) Validate(userName string, auth string) bool {
	if !v.enable {
		return true
	}
	secret, ok := v.otpSecretMap[userName]
	if ok && totp.Validate(auth, secret) {
		log.Infof(v.logHeader, "validate totp for user %v success", userName)
		return true
	}
	log.Infof(v.logHeader, "validate totp for user %v fail", userName)
	return false
}
