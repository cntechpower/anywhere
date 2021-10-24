package handler

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	log "github.com/cntechpower/utils/log.v2"
)

func SessionFilter(c *gin.Context) {
	fields := map[string]interface{}{
		log.FieldNameBizName: "http.handler.sessionFilter",
	}
	if strings.HasPrefix(c.Request.URL.Path, "/static/") {
		c.Next()
		return
	}
	if c.Request.URL.Path == "/user/login" ||
		c.Request.URL.Path == "/user_login" ||
		c.Request.URL.Path == "/report" ||
		c.Request.URL.Path == "/metrics" ||
		c.Request.URL.Path == "/manifest.json" {
		c.Next()
		return
	}
	session := sessions.Default(c)
	authHeader := session.Get("auth")
	tokenString, ok := authHeader.(string)
	if !ok {
		log.Warnf(fields, "get empty auth")
		redirectToLogin(c)
		return
	}

	if !jwtValidator.Validate("", tokenString) {
		log.Warnf(fields, "validate jwt for %s fail", c.ClientIP())
		redirectToLogin(c)
		return
	}
}

func UserLogin(c *gin.Context) {
	//get username/password/otpcode from form
	session := sessions.Default(c)
	userName, ok := c.GetPostForm("username")
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, RespUserPassIsRequired)
		return
	}
	password, ok := c.GetPostForm("password")
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, RespUserPassIsRequired)
		return
	}
	otpCode, ok := c.GetPostForm("otpcode")
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, RespUserPassIsRequired)
		return
	}

	if !userValidator.Validate(userName, password, otpCode) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, RespUserPassWrong)
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
	c.JSON(http.StatusOK, RespUserLoginSuccess)
	c.Next()
	return
}

func userLogout(c *gin.Context) {
}
