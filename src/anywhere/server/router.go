package main

import (
	"anywhere/log"
	"anywhere/server/restapi/api/restapi"
	"anywhere/server/restapi/api/restapi/operations"
	"anywhere/util"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"

	"github.com/go-openapi/loads"

	"github.com/gin-gonic/gin"
)

var jwtKey = []byte("anywhereToken")
var initUser = "aROnOCXRQZBx5vNT"
var initPass = "P8xw8RCxBCm7Holh"

var (
	ErrUserPassIsRequired = gin.H{"message": "username is required"}
	ErrUserPassWrong      = gin.H{"message": "username/password wrong"}
)

func getJwt(user string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": user,
		"nbf":  time.Now(),
		"exp":  time.Now().Add(8 * time.Hour),
	})
	return token.SignedString(jwtKey)
}

func checkJwt(c *gin.Context) error {
	session := sessions.Default(c)
	auth := session.Get("auth")
	tokenString, ok := auth.(string)
	if !ok {
		redirectToLogin(c)
	}
	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (i interface{}, e error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})
	return err
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
	react.Any("/user/*any", renderIndex)
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

func redirectToLogin(c *gin.Context) {
	c.Redirect(http.StatusTemporaryRedirect, "/react/user/login")
	c.Abort()
}

func sessionFilter(c *gin.Context) {
	l := log.GetCustomLogger("sessionFilter")
	l.Infof("request path: %v", c.Request.URL.Path)
	if strings.HasPrefix(c.Request.URL.Path, "/react/static/") {
		c.Next()
		return
	}
	if c.Request.URL.Path == "/react/user/login" || c.Request.URL.Path == "/user_login" {
		c.Next()
		return
	}

	if err := checkJwt(c); err != nil {
		l.Infof("check jwt error: %v", err)
		redirectToLogin(c)
	}
}

func userLogin(c *gin.Context) {
	session := sessions.Default(c)
	userName, ok := c.GetPostForm("username")
	log.GetDefaultLogger().Infof("called userLogin")
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrUserPassIsRequired)
		return
	}
	password, ok := c.GetPostForm("password")
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrUserPassIsRequired)
		return
	}
	if userName != initUser || password != initPass {
		c.AbortWithStatusJSON(http.StatusNotAcceptable, ErrUserPassWrong)
		return
	}
	log.GetDefaultLogger().Infof("called userLogin")
	token, err := getJwt(userName)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	session.Set("auth", token)
	log.GetDefaultLogger().Info(session.Save())
	c.Header("Access-Control-Allow-Origin", "*")
	c.JSON(http.StatusOK, gin.H{"message": "login success"})
	//c.Redirect(http.StatusTemporaryRedirect, "/react/")
}

func startUIAndAPIService(addr, user, pass string, errChan chan error) {
	if err := util.CheckAddrValid(addr); err != nil {
		errChan <- err
	}
	router := gin.New()
	initUser = user
	initPass = pass
	//session auth
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("anywhere", store))
	router.Use(sessionFilter)

	//router.LoadHTMLFiles("./static/login.html", "./static/index.html")
	router.LoadHTMLFiles("./static/index.html")
	//router.Any("/login", func(c *gin.Context) {
	//	c.HTML(http.StatusOK, "login.html", nil)
	//})
	router.POST("/user_login", userLogin)

	//header auth
	//router.Use(authFilter)

	if err := addUIRouter(router); err != nil {
		errChan <- err
	}
	if err := addAPIRouter(router); err != nil {
		errChan <- err
	}
	//not support tls
	//http://10.0.0.8/self-code/anywhere/issues/29
	//if certFile != "" && keyFile != "" {
	//	errChan <- router.RunTLS(addr, certFile, keyFile)
	//}
	errChan <- router.Run(addr)

}
