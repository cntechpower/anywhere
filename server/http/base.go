package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func redirectToLogin(c *gin.Context) {
	c.Redirect(http.StatusTemporaryRedirect, "/user/login")
	c.Abort()
}

func rejectNoLogin(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, NewResp(http.StatusUnauthorized, "未登录"))
	c.Abort()
}

func NewResp(code int64, data interface{}) gin.H {
	return gin.H{
		"code": code,
		"data": data,
	}
}
