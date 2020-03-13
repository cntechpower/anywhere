package auth

import (
	"anywhere/log"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/dgrijalva/jwt-go"
)

type JwtValidator struct {
	jwtKey []byte
	logger *logrus.Entry
}

func NewJwtValidator() *JwtValidator {
	return &JwtValidator{
		jwtKey: []byte("anywhere"),
		logger: log.GetCustomLogger("jwtValidator")}
}

func (v *JwtValidator) Generate(userName string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": userName,
		"nbf":  time.Now(),
		"exp":  time.Now().Add(8 * time.Hour),
	})
	v.logger.Infof("generate jwt for user %v", userName)
	return token.SignedString(v.jwtKey)
}

func (v *JwtValidator) Validate(userName string, auth string) bool {
	_, err := jwt.Parse(auth, func(token *jwt.Token) (i interface{}, e error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.jwtKey, nil
	})
	if err != nil {
		v.logger.Infof("validate jwt for user %v fail", userName)
		return false
	}
	v.logger.Infof("validate jwt for user %v success", userName)
	return true
}
