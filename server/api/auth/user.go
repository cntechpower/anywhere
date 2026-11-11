package auth

import (
	"crypto/subtle"

	"github.com/cntechpower/utils/log"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"

	"github.com/cntechpower/anywhere/model"
)

// UserValidator authenticates users from the server configuration.
//
// Passwords can be stored either as a bcrypt hash (user_password_hash, the
// recommended form) or as plain text (user_password, legacy). Plain text
// passwords are still accepted so existing deployments keep working, but a
// warning is emitted at startup.
type UserValidator struct {
	userPassMap map[string]*model.User
	logHeader   *log.Header
	dummyHash   []byte
}

func NewUserValidator(users *model.UserConfig) *UserValidator {
	v := &UserValidator{
		userPassMap: make(map[string]*model.User),
		logHeader:   log.NewHeader("UserValidator"),
	}
	// Used to keep the failure path roughly constant-time for unknown users.
	if h, err := bcrypt.GenerateFromPassword([]byte("anywhere-dummy-password"), bcrypt.DefaultCost); err == nil {
		v.dummyHash = h
	}
	if users == nil {
		return v
	}
	for _, user := range users.Users {
		if user == nil || user.UserName == "" {
			continue
		}
		v.userPassMap[user.UserName] = user
		if user.UserPassHash == "" {
			log.Warnf(v.logHeader, "user %q uses a plain text password, please switch to user_password_hash", user.UserName)
		} else if user.UserPass != "" {
			log.Warnf(v.logHeader, "user %q defines both user_password and user_password_hash, the hash wins", user.UserName)
		}
	}
	return v
}

// ValidateUserPass checks the username/password pair. It never short-circuits
// before a bcrypt comparison so response time does not reveal whether a user
// exists.
func (v *UserValidator) ValidateUserPass(userName, password string) bool {
	user, ok := v.userPassMap[userName]
	if !ok {
		if v.dummyHash != nil {
			_ = bcrypt.CompareHashAndPassword(v.dummyHash, []byte(password))
		}
		return false
	}
	if user.UserPassHash != "" {
		return bcrypt.CompareHashAndPassword([]byte(user.UserPassHash), []byte(password)) == nil
	}
	if user.UserPass == "" || password == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(user.UserPass), []byte(password)) == 1
}

// ValidateOtp returns true when OTP is disabled for the user, otherwise it
// validates the TOTP code.
func (v *UserValidator) ValidateOtp(userName, otpCode string) bool {
	user, ok := v.userPassMap[userName]
	if !ok {
		return false
	}
	if !user.OtpEnable {
		return true
	}
	if otpCode == "" {
		return false
	}
	return totp.Validate(otpCode, user.OtpCode)
}

// Authenticate validates both the password and (when enabled) the OTP code.
// Both factors are required; skipping the password is no longer possible.
func (v *UserValidator) Authenticate(userName, password, otpCode string) bool {
	if !v.ValidateUserPass(userName, password) {
		log.Infof(v.logHeader, "password validation for user %q failed", userName)
		return false
	}
	if !v.ValidateOtp(userName, otpCode) {
		log.Infof(v.logHeader, "otp validation for user %q failed", userName)
		return false
	}
	return true
}
