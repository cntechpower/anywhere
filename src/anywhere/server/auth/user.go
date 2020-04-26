package auth

import (
	"anywhere/log"
	"anywhere/util"
)

type UserValidator struct {
	userPassMap map[string] /*userName*/ string /*password*/
	logHeader   *log.Header
}

func NewUserValidator(userName string, password string) *UserValidator {
	return &UserValidator{
		userPassMap: map[string]string{
			userName: password,
		},
		logHeader: log.NewHeader("UserValidator"),
	}
}

func (v *UserValidator) Generate(userName string) (string, error) {
	randPass := util.RandString(16)
	v.userPassMap[userName] = randPass
	log.Infof(v.logHeader, "generate password for user %v", userName)
	return randPass, nil
}

func (v *UserValidator) Validate(userName string, auth string) bool {
	pass, ok := v.userPassMap[userName]
	if ok && (auth == pass) {
		log.Infof(v.logHeader, "validate password for user %v success", userName)
		return true
	}
	log.Infof(v.logHeader, "validate password for user %v fail", userName)
	return false
}
