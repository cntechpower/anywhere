package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func redirectToLogin(c *gin.Context) {
	c.Redirect(http.StatusTemporaryRedirect, "/user/login")
	c.Abort()
}

func NewResp(code int64, data interface{}) gin.H {
	return gin.H{
		"code": code,
		"data": data,
	}
}
