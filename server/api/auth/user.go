package auth

import (
	"github.com/cntechpower/anywhere/model"
	log "github.com/cntechpower/utils/log.v2"

	"github.com/pquerna/otp/totp"
)

type UserValidator struct {
	userPassMap map[string] /*userName*/ *model.User /*password*/
}

func NewUserValidator(users *model.UserConfig) *UserValidator {
	u := &UserValidator{
		userPassMap: make(map[string]*model.User, len(users.Users)),
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
	fields := map[string]interface{}{
		log.FieldNameBizName: "UserValidator.ValidateUserPass",
	}
	user, ok := v.userPassMap[userName]
	if !ok || (auth != user.UserPass) {
		log.Infof(fields, "validate password for user %v fail", userName)
		return false
	}
	log.Infof(fields, "validate password for user %v success", userName)
	return true
}

func (v *UserValidator) ValidateOtp(userName, otpCode string) bool {
	fields := map[string]interface{}{
		log.FieldNameBizName: "UserValidator.ValidateOtp",
	}
	user, ok := v.userPassMap[userName]
	if !ok || (user.OtpEnable && !totp.Validate(otpCode, user.OtpCode)) {
		log.Infof(fields, "validate totp for user %v fail", userName)
		return false
	}
	log.Infof(fields, "validate totp for user %v success", userName)
	return true
}
