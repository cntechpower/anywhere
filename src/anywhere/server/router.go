package main

import (
	"anywhere/log"
	"anywhere/server/restapi/api/restapi"
	"anywhere/server/restapi/api/restapi/operations"
	"anywhere/util"
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"

	"github.com/go-openapi/loads"

	"github.com/gin-gonic/gin"
)

var secrets = gin.H{
	"dujinyang": gin.H{"email": "dujinyang@cntechpower.com", "phone": "13681611995"},
}

func addUIRouter(router *gin.Engine) error {
	if !util.CheckPathExist("./static") {
		return fmt.Errorf("static dir not found")
	}
	react := router.Group("/react/")

	renderIndex := func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	}
	react.Any("/", renderIndex)
	react.Any("/proxy/*any", renderIndex)
	react.Any("/note/*any", renderIndex)
	react.StaticFS("/static/", http.Dir("./static/static"))
	react.StaticFile("/manifest.json", "./static/manifest.json")
	react.StaticFile("/logo192.png", "./static/logo192.png")
	return nil
}

func addAPIRouter(router *gin.Engine) error {
	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		return err
	}

	api := operations.NewAnywhereServerAPI(swaggerSpec)
	server := restapi.NewServer(api)
	defer server.Shutdown()
	server.Port = port
	server.ConfigureAPI()
	apiRouter := router.Group("/api")
	handler := server.GetHandler()
	apiRouter.Any("/*any", gin.WrapH(handler))
	return nil
}

func authFilter(c *gin.Context) {
	l := log.GetCustomLogger("gin")
	l.Infof("request path: %v", c.Request.URL.Path)
	if c.Request.URL.Path == "/login" || c.Request.URL.Path == "/user_login" {
		return
	}
	authorization := c.GetHeader("authorization")

	if authorization == "" {
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		c.Abort()
	}
}

func sessionFilter(c *gin.Context) {
	l := log.GetCustomLogger("sessionFilter")
	l.Infof("request path: %v", c.Request.URL.Path)
	if c.Request.URL.Path == "/login" || c.Request.URL.Path == "/user_login" {
		return
	}
	session := sessions.Default(c)

	auth := session.Get("auth")
	l.Infof("auth: %v", auth)
	if auth == nil || auth.(string) == "" {
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		c.Abort()
	}
}

func userLogin(c *gin.Context) {
	session := sessions.Default(c)
	userName, ok := c.GetPostForm("username")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "username is required"})
		return
	}
	password, ok := c.GetPostForm("password")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "password is required"})
		return
	}
	if userName != "admin" || password != "admin" {
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"message": "username/password wrong"})
	}
	log.GetDefaultLogger().Infof("called userLogin")
	session.Set("auth", userName+password)
	log.GetDefaultLogger().Info(session.Save())
	c.Redirect(http.StatusTemporaryRedirect, "/react/")
}

func startUIAndAPIService(addr, certFile, keyFile string, errChan chan error) {
	if err := util.CheckAddrValid(addr); err != nil {
		errChan <- err
	}
	router := gin.New()

	//session auth
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("anywhere", store))
	router.Use(sessionFilter)

	router.LoadHTMLFiles("./static/login.html", "./static/index.html")
	router.Any("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})
	router.POST("/user_login", userLogin)

	//header auth
	//router.Use(authFilter)

	if err := addUIRouter(router); err != nil {
		errChan <- err
	}
	if err := addAPIRouter(router); err != nil {
		errChan <- err
	}
	//TODO: tls
	//if certFile != "" && keyFile != "" {
	//	errChan <- router.RunTLS(addr, certFile, keyFile)
	//}
	errChan <- router.Run(addr)

}
