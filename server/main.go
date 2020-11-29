package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cntechpower/anywhere/log"
	"github.com/cntechpower/anywhere/server/anywhereServer"
	"github.com/cntechpower/anywhere/server/auth"
	"github.com/cntechpower/anywhere/server/cmd"
	"github.com/cntechpower/anywhere/server/handler/rpcHandler"
	"github.com/cntechpower/anywhere/server/restapi/api/restapi"
	"github.com/cntechpower/anywhere/server/restapi/api/restapi/operations"
	"github.com/cntechpower/anywhere/tls"
	"github.com/cntechpower/anywhere/util"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/go-openapi/loads"

	"github.com/spf13/cobra"
)

//server global config
var version string

func main() {
	log.InitLogger("")
	var rootCmd = &cobra.Command{
		Use:   "anywhered",
		Short: "This is A Proxy Server ",
		Long:  "anywhere server - " + version,
	}
	var startCmd = &cobra.Command{
		Use:   "start",
		Short: "start anywhered service",
		Long:  "anywhere server Version 0.0.1 -" + version,
		Run: func(cmd *cobra.Command, args []string) {
			if err := run(cmd, args); err != nil {
				log.Fatalf(log.NewHeader("serverMain"), err.Error())
			}
		},
	}

	//main service
	rootCmd.AddCommand(startCmd)
	//agent cmds
	rootCmd.AddCommand(cmd.GetAgentCmd())

	//proxy cmds
	rootCmd.AddCommand(cmd.GetProxyCmd())

	//config file manage cmds
	rootCmd.AddCommand(cmd.GetConfigCmd())

	//conn cmds
	rootCmd.AddCommand(cmd.GetConnCmd())
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}

}

func run(_ *cobra.Command, _ []string) error {
	h := log.NewHeader("serverMain")
	c, err := anywhereServer.ParseSystemConfigFile()
	if err != nil {
		return err
	}
	s := anywhereServer.InitServerInstance(c.ServerId, c.MainPort, c.User)
	tlsConfig, err := tls.ParseTlsConfig(c.Ssl.CertFile, c.Ssl.KeyFile, c.Ssl.CaFile)
	if err != nil {
		return err
	}
	s.SetCredentials(tlsConfig)

	//start main service
	s.Start()

	// start rpc server
	rpcExitChan := make(chan error, 0)
	go rpcHandler.StartRpcServer(s, c.UiConfig.GrpcAddr, rpcExitChan)
	webExitChan := make(chan error, 0)
	if c.UiConfig.IsWebEnable {
		go startUIAndAPIService(c.UiConfig.WebAddr, webExitChan, c.UiConfig.SkipLogin, c.UiConfig.DebugMode)

	}

	//wait for os kill signal. TODO: graceful shutdown
	go util.ListenTTINSignalLoop()
	serverExitChan := util.ListenKillSignal()
	select {
	case <-serverExitChan:
		log.Infof(h, "Server Existing")
	case err := <-webExitChan:
		log.Fatalf(h, "api server exit with error: %v", err)
	case err := <-rpcExitChan:
		log.Fatalf(h, "rpc server exit with error: %v", err)
	case err := <-s.ExitChan:
		log.Fatalf(h, "anywhere server exit with error: %v", err)
	}
	return nil
}

var userValidator *auth.UserValidator
var jwtValidator *auth.JwtValidator

var (
	ErrUserPassIsRequired = gin.H{"message": "username/password/otp_code is required"}
	ErrUserPassWrong      = gin.H{"message": "username/password wrong"}
)

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
	h := log.NewHeader("sessionFilter")
	if strings.HasPrefix(c.Request.URL.Path, "/react/static/") {
		c.Next()
		return
	}
	if c.Request.URL.Path == "/react/user/login" || c.Request.URL.Path == "/user_login" {
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

func startUIAndAPIService(addr string, errChan chan error, skipLogin, debug bool) {
	if err := util.CheckAddrValid(addr); err != nil {
		errChan <- err
	}
	router := gin.New()
	if debug {
		// running in debug mode, open access log
		gin.SetMode(gin.DebugMode)
		router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
			Formatter: func(param gin.LogFormatterParams) string {
				return fmt.Sprintf("[%s] %s \"%s %s %s %d %s \"%s\" %s\"\n",
					param.TimeStamp.Format(time.RFC3339),
					param.ClientIP,
					param.Method,
					param.Path,
					param.Request.Proto,
					param.StatusCode,
					param.Latency,
					param.Request.UserAgent(),
					param.ErrorMessage,
				)
			},
			Output:    nil,
			SkipPaths: []string{"/react/static"},
		}))
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	userValidator = anywhereServer.GetServerInstance().GetUserValidator()
	jwtValidator = auth.NewJwtValidator()
	//session auth
	store := cookie.NewStore([]byte(util.RandString(16)))
	router.Use(sessions.Sessions("anywhere", store))
	//support frontend development
	if !skipLogin {
		router.Use(sessionFilter)
	}

	router.LoadHTMLFiles("./static/index.html")
	router.POST("/user_login", userLogin)
	if err := addUIRouter(router); err != nil {
		errChan <- err
	}
	if err := addAPIRouter(router); err != nil {
		errChan <- err
	}
	errChan <- router.Run(addr)

}
