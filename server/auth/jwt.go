package auth

import (
	"fmt"
	"time"

	"github.com/cntechpower/anywhere/log"

	"github.com/dgrijalva/jwt-go"
)

type JwtValidator struct {
	jwtKey    []byte
	logHeader *log.Header
}

func NewJwtValidator() *JwtValidator {
	return &JwtValidator{
		jwtKey:    []byte("anywhere"),
		logHeader: log.NewHeader("JwtValidator"),
	}
}

func (v *JwtValidator) Generate(userName string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": userName,
		"nbf":  time.Now(),
		"exp":  time.Now().Add(8 * time.Hour),
	})
	log.Infof(v.logHeader, "generate jwt for user %v", userName)
	return token.SignedString(v.jwtKey)
}

func (v *JwtValidator) Validate(userName string, auth string) bool {
	if auth == "" {
		log.Infof(v.logHeader, "validate jwt fail because jwt is empty")
		return false
	}
	_, err := jwt.Parse(auth, func(token *jwt.Token) (i interface{}, e error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.jwtKey, nil
	})
	if err != nil {
		return false
	}
	return true
}
