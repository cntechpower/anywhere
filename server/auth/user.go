package auth

import (
	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/utils/log"

	"github.com/pquerna/otp/totp"
)

type UserValidator struct {
	userPassMap map[string] /*userName*/ *model.User /*password*/
	logHeader   *log.Header
}

func NewUserValidator(users *model.UserConfig) *UserValidator {
	u := &UserValidator{
		userPassMap: make(map[string]*model.User, len(users.Users)),
		logHeader:   log.NewHeader("UserValidator"),
	}
	for _, user := range users.Users {
		u.userPassMap[user.UserName] = user
	}
	return u
}

func (v *UserValidator) Validate(userName, password, otpCode string) bool {
	return v.ValidateOtp(userName, otpCode) && v.ValidateUserPass(userName, password)
}

func (v *UserValidator) ValidateUserPass(userName, auth string) bool {
	user, ok := v.userPassMap[userName]
	if !ok || (auth != user.UserPass) {
		log.Infof(v.logHeader, "validate password for user %v fail", userName)
		return false
	}
	log.Infof(v.logHeader, "validate password for user %v success", userName)
	return true
}

func (v *UserValidator) ValidateOtp(userName, otpCode string) bool {
	user, ok := v.userPassMap[userName]
	if !ok || (user.OtpEnable && !totp.Validate(user.OtpCode, otpCode)) {
		log.Infof(v.logHeader, "validate totp for user %v fail", userName)
		return false
	}
	log.Infof(v.logHeader, "validate totp for user %v success", userName)
	return true
}
