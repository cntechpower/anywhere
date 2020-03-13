package auth

import (
	"anywhere/log"
	"anywhere/util"

	"github.com/sirupsen/logrus"
)

type UserValidator struct {
	userPassMap map[string] /*userName*/ string /*password*/
	logger      *logrus.Entry
}

func NewUserValidator(userName string, password string) *UserValidator {
	return &UserValidator{
		userPassMap: map[string]string{
			userName: password,
		},
		logger: log.GetCustomLogger("userValidator"),
	}
}

func (v *UserValidator) Generate(userName string) (string, error) {
	randPass := util.RandString(16)
	v.userPassMap[userName] = randPass
	v.logger.Infof("generate password for user %v", userName)
	return randPass, nil
}

func (v *UserValidator) Validate(userName string, auth string) bool {
	pass, ok := v.userPassMap[userName]
	if ok && (auth == pass) {
		v.logger.Infof("validate password for user %v success", userName)
		return true
	}
	v.logger.Infof("validate password for user %v fail", userName)
	return false
}
