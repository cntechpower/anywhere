package auth

import (
	"fmt"
	"time"

	"github.com/cntechpower/utils/log"
	"github.com/golang-jwt/jwt/v5"
)

const (
	// minJwtKeyLength rejects trivially guessable signing keys.
	minJwtKeyLength = 16
	// tokenTTL is how long a UI session token stays valid.
	tokenTTL = 7 * 24 * time.Hour
)

type JwtValidator struct {
	jwtKey    []byte
	logHeader *log.Header
	ttl       time.Duration
}

func NewJwtValidator(key []byte) (*JwtValidator, error) {
	if len(key) < minJwtKeyLength {
		return nil, fmt.Errorf("jwt secret must be at least %d bytes, got %d", minJwtKeyLength, len(key))
	}
	return &JwtValidator{
		jwtKey:    key,
		logHeader: log.NewHeader("JwtValidator"),
		ttl:       tokenTTL,
	}, nil
}

func (v *JwtValidator) Generate(userName string) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": userName,
		"iat":  now.Unix(),
		"nbf":  now.Unix(),
		"exp":  now.Add(v.ttl).Unix(),
	})
	tokenString, err := token.SignedString(v.jwtKey)
	if err != nil {
		return "", err
	}
	log.Infof(v.logHeader, "generated jwt for user %q", userName)
	return tokenString, nil
}

func (v *JwtValidator) Validate(userName, tokenString string) bool {
	if tokenString == "" {
		return false
	}
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.jwtKey, nil
	})
	if err != nil {
		log.Infof(v.logHeader, "parse jwt error: %v", err)
		return false
	}
	if !token.Valid {
		return false
	}
	userNameI, ok := claims["user"]
	if !ok {
		return false
	}
	userNameStr, ok := userNameI.(string)
	if !ok || userNameStr == "" {
		return false
	}
	if userName != "" && userNameStr != userName {
		log.Infof(v.logHeader, "expected user %q, got %q", userName, userNameStr)
		return false
	}
	return true
}
