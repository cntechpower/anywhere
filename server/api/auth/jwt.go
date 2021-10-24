package auth

import (
	"fmt"
	"time"

	"github.com/cntechpower/utils/log"

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
		//Not Before
		"nbf": time.Now().Unix(),
		//Expire Time
		"exp": time.Now().Add(24 * time.Hour * 14).Unix(),
	})
	log.Infof(v.logHeader, "generate jwt for user %v", userName)
	return token.SignedString(v.jwtKey)
}

func (v *JwtValidator) Validate(userName string, auth string) bool {
	if auth == "" {
		v.logHeader.Errorf("validate jwt fail because jwt is empty")
		return false
	}
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(auth, claims, func(token *jwt.Token) (i interface{}, e error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.jwtKey, nil
	})
	if err != nil {
		v.logHeader.Errorf("parse jwt with claims error: %v", err)
		return false
	}

	//check userName
	if userName != "" {
		userNameI := claims["user"]
		if userNameI == nil {
			v.logHeader.Errorf("expected userName=%v, got nil", userName)
			return false
		}

		userNameStr, ok := userNameI.(string)
		if !ok {
			v.logHeader.Errorf("expected userName=%v, got empty", userName)
			return false
		}
		if userNameStr != userName {
			v.logHeader.Errorf("expected userName=%v, got %v", userName, userNameStr)
			return false
		}
	}

	return true
}
