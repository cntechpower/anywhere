package handler

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cntechpower/utils/log"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	loginMaxFailures    = 10
	loginFailureWindow  = 5 * time.Minute
	apiUnauthorizedCode = 401
)

type failureRecord struct {
	count int
	first time.Time
}

var loginFailures = struct {
	sync.Mutex
	m map[string]*failureRecord
}{m: make(map[string]*failureRecord)}

func allowLogin(ip string) bool {
	loginFailures.Lock()
	defer loginFailures.Unlock()
	r, ok := loginFailures.m[ip]
	if !ok {
		return true
	}
	if time.Since(r.first) > loginFailureWindow {
		delete(loginFailures.m, ip)
		return true
	}
	return r.count < loginMaxFailures
}

func recordLoginFailure(ip string) {
	loginFailures.Lock()
	defer loginFailures.Unlock()
	if len(loginFailures.m) > 10000 {
		// Bound the map under a distributed brute force attempt.
		loginFailures.m = make(map[string]*failureRecord)
	}
	r, ok := loginFailures.m[ip]
	if !ok || time.Since(r.first) > loginFailureWindow {
		loginFailures.m[ip] = &failureRecord{count: 1, first: time.Now()}
		return
	}
	r.count++
}

func clearLoginFailures(ip string) {
	loginFailures.Lock()
	defer loginFailures.Unlock()
	delete(loginFailures.m, ip)
}

// isAuthExempt lists the only paths that may be reached without a session.
func isAuthExempt(path string) bool {
	switch path {
	case "/user/login", "/user_login", "/user_logout",
		"/report", "/metrics", "/manifest.json", "/favicon.ico":
		return true
	}
	return strings.HasPrefix(path, "/static/") || strings.HasPrefix(path, "/assets/")
}

// SessionFilter requires a valid session token. API calls get a JSON 401 so
// the UI can react, page requests are redirected to the login page.
func (a *API) SessionFilter(c *gin.Context) {
	h := log.NewHeader("sessionFilter")
	path := c.Request.URL.Path
	if isAuthExempt(path) {
		c.Next()
		return
	}

	session := sessions.Default(c)
	tokenString, _ := session.Get("auth").(string)
	if tokenString == "" || !a.jwt.Validate("", tokenString) {
		h.Warnf("unauthorized request to %s from %s", path, c.ClientIP())
		if strings.HasPrefix(path, "/api/") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": apiUnauthorizedCode,
				"data": "session expired, please login again",
			})
			return
		}
		redirectToLogin(c)
		return
	}
	c.Next()
}

// UserLogin authenticates username + password and, when enabled for the user,
// a TOTP code. Both factors are mandatory.
func (a *API) UserLogin(c *gin.Context) {
	ip := c.ClientIP()
	if !allowLogin(ip) {
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
			"code": http.StatusTooManyRequests,
			"data": "too many login attempts, please retry later",
		})
		return
	}

	userName := strings.TrimSpace(c.PostForm("username"))
	password := c.PostForm("password")
	otpCode := c.PostForm("otpcode")
	if userName == "" || password == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, RespUserPassIsRequired)
		return
	}

	if !a.users.Authenticate(userName, password, otpCode) {
		recordLoginFailure(ip)
		c.AbortWithStatusJSON(http.StatusUnauthorized, RespUserPassWrong)
		return
	}
	clearLoginFailures(ip)

	token, err := a.jwt.Generate(userName)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	session := sessions.Default(c)
	session.Set("auth", token)
	if err := session.Save(); err != nil {
		log.NewHeader("UserLogin").Errorf("save session error: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, RespUserLoginSuccess)
}

// UserLogout clears the session cookie.
func (a *API) UserLogout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	if err := session.Save(); err != nil {
		log.NewHeader("UserLogout").Errorf("save session error: %v", err)
	}
	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "data": "logout success"})
}
