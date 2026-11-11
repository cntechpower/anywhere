package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"

	"github.com/cntechpower/anywhere/model"
)

const testOtpSecret = "ZKQVBFY55NJGGWBV5F6CU5CEK2YAWIB6"

func newTestValidator(t *testing.T) *UserValidator {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("s3cret"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	return NewUserValidator(&model.UserConfig{Users: []*model.User{
		{
			UserName:  "plain",
			UserPass:  "plain-pass",
			OtpEnable: false,
		},
		{
			UserName:     "otp",
			UserPassHash: string(hash),
			OtpEnable:    true,
			OtpCode:      testOtpSecret,
		},
		{
			UserName:  "nopass",
			OtpEnable: false,
		},
	}})
}

// TestAuthenticateRequiresPassword is the regression test for the login
// bypass: previously only the OTP code was checked, so any password (or none)
// granted a session when OTP was disabled.
func TestAuthenticateRequiresPassword(t *testing.T) {
	v := newTestValidator(t)

	if v.Authenticate("plain", "", "") {
		t.Fatal("empty password must not authenticate")
	}
	if v.Authenticate("plain", "wrong", "") {
		t.Fatal("wrong password must not authenticate")
	}
	if !v.Authenticate("plain", "plain-pass", "") {
		t.Fatal("correct password with OTP disabled must authenticate")
	}
	if v.Authenticate("nopass", "", "") {
		t.Fatal("user without password must never authenticate")
	}
	if v.Authenticate("unknown", "plain-pass", "") {
		t.Fatal("unknown user must not authenticate")
	}
}

func TestAuthenticateWithOtp(t *testing.T) {
	v := newTestValidator(t)
	code, err := totp.GenerateCode(testOtpSecret, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	if !v.Authenticate("otp", "s3cret", code) {
		t.Fatal("expected password + valid otp to authenticate")
	}
	if v.Authenticate("otp", "s3cret", "") {
		t.Fatal("missing otp must not authenticate when enabled")
	}
	if v.Authenticate("otp", "s3cret", "000000") {
		t.Fatal("wrong otp must not authenticate")
	}
	if v.Authenticate("otp", "wrong", code) {
		t.Fatal("wrong password must not authenticate even with valid otp")
	}
}

func TestValidateUserPassConstantTime(t *testing.T) {
	v := newTestValidator(t)
	if !v.ValidateUserPass("plain", "plain-pass") {
		t.Fatal("expected legacy plaintext password to validate")
	}
	if v.ValidateUserPass("plain", "plain-pas") {
		t.Fatal("prefix password must not validate")
	}
	if !v.ValidateUserPass("otp", "s3cret") {
		t.Fatal("expected bcrypt password to validate")
	}
}

func newTestJwtValidator(t *testing.T) *JwtValidator {
	t.Helper()
	v, err := NewJwtValidator([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestJwtRoundTrip(t *testing.T) {
	v := newTestJwtValidator(t)
	token, err := v.Generate("alice")
	if err != nil {
		t.Fatal(err)
	}
	if !v.Validate("alice", token) {
		t.Fatal("expected token to validate")
	}
	if v.Validate("bob", token) {
		t.Fatal("token must not validate for another user")
	}
}

func TestJwtRejectsForgedSignature(t *testing.T) {
	v := newTestJwtValidator(t)
	forged, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": "alice",
		"exp":  time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte("a-different-secret-key-1234"))
	if err != nil {
		t.Fatal(err)
	}
	if v.Validate("alice", forged) {
		t.Fatal("token signed with a different key must be rejected")
	}
}

func TestJwtRejectsExpiredToken(t *testing.T) {
	v := newTestJwtValidator(t)
	expired, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": "alice",
		"exp":  time.Now().Add(-time.Hour).Unix(),
	}).SignedString([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	if v.Validate("alice", expired) {
		t.Fatal("expired token must be rejected")
	}
}

func TestJwtRejectsAlgorithmConfusion(t *testing.T) {
	v := newTestJwtValidator(t)
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS384, jwt.MapClaims{
		"user": "alice",
		"exp":  time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	if v.Validate("alice", token) {
		t.Fatal("only HS256 must be accepted")
	}
}

func TestJwtSecretTooShort(t *testing.T) {
	if _, err := NewJwtValidator([]byte("short")); err == nil {
		t.Fatal("expected error for short secret")
	}
}
