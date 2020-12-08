package http

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/cntechpower/anywhere/log"
)

func redirectToLogin(c *gin.Context) {
	c.Redirect(http.StatusTemporaryRedirect, "/user/login")
	c.Abort()
}

func sessionFilter(c *gin.Context) {
	h := log.NewHeader("sessionFilter")
	if strings.HasPrefix(c.Request.URL.Path, "/static/") {
		c.Next()
		return
	}
	if c.Request.URL.Path == "/user/login" ||
		c.Request.URL.Path == "/user_login" ||
		c.Request.URL.Path == "/report" ||
		c.Request.URL.Path == "/metrics" {
		c.Next()
		return
	}
	session := sessions.Default(c)
	authHeader := session.Get("auth")
	tokenString, ok := authHeader.(string)
	if !ok {
		redirectToLogin(c)
	}

	if !jwtValidator.Validate("", tokenString) {
		log.Warnf(h, "validate jwt for %s fail", c.ClientIP())
		redirectToLogin(c)
	}
}

func userLogin(c *gin.Context) {
	//get username/password/otpcode from form
	session := sessions.Default(c)
	userName, ok := c.GetPostForm("username")
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrUserPassIsRequired)
		return
	}
	password, ok := c.GetPostForm("password")
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrUserPassIsRequired)
		return
	}
	otpCode, ok := c.GetPostForm("otpcode")
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrUserPassIsRequired)
		return
	}

	if !userValidator.Validate(userName, password, otpCode) {
		c.AbortWithStatusJSON(http.StatusNotAcceptable, ErrUserPassWrong)
		return
	}

	//validate success, generate setCookie
	token, err := jwtValidator.Generate(userName)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	session.Set("auth", token)
	_ = session.Save() //ignore session save error
	c.Header("Access-Control-Allow-Origin", "*")
	c.JSON(http.StatusOK, gin.H{"message": "login success"})
}
