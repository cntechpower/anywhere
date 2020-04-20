package auth

import (
	"anywhere/log"

	"github.com/pquerna/otp/totp"
)

type TOTPValidator struct {
	otpSecretMap map[string] /*userName*/ string /*secret*/
	logger       *log.Logger
	enable       bool
}

func NewTOTPValidator(adminUser, adminSecret string, enable bool) *TOTPValidator {
	return &TOTPValidator{
		otpSecretMap: map[string]string{
			adminUser: adminSecret,
		},
		logger: log.GetDefaultLogger(),
		enable: enable,
	}
}

func (v *TOTPValidator) Generate(userName string) (string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "cntechpower.com",
		AccountName: userName,
	})
	if err != nil {
		v.logger.Infof("generate totp secret for user %v fail", userName)
		return "", err
	}
	v.logger.Infof("generate totp secret for user %v success", userName)
	v.otpSecretMap[userName] = key.Secret()
	return key.Secret(), nil
}

func (v *TOTPValidator) Validate(userName string, auth string) bool {
	if !v.enable {
		return true
	}
	secret, ok := v.otpSecretMap[userName]
	if ok && totp.Validate(auth, secret) {
		v.logger.Infof("validate totp for user %v success", userName)
		return true
	}
	v.logger.Infof("validate totp for user %v fail", userName)
	return false
}
