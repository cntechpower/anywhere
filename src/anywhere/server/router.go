package main

import (
	"anywhere/log"
	"anywhere/server/auth"
	"anywhere/server/restapi/api/restapi"
	"anywhere/server/restapi/api/restapi/operations"
	"anywhere/util"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"

	"github.com/go-openapi/loads"

	"github.com/gin-gonic/gin"
)

var userValidator *auth.UserValidator
var jwtValidator *auth.JwtValidator
var totpValidator *auth.TOTPValidator

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
		log.Warnf("validate jwt for %s fail", c.ClientIP())
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

	//username & password check
	if !userValidator.Validate(userName, password) {
		c.AbortWithStatusJSON(http.StatusNotAcceptable, ErrUserPassWrong)
		return
	}

	//OTP check
	if !totpValidator.Validate(userName, otpCode) {
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

func startUIAndAPIService(addr, user, pass, totpSecret string, otpEnable bool, errChan chan error, skipLogin bool) {
	if err := util.CheckAddrValid(addr); err != nil {
		errChan <- err
	}
	router := gin.New()
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
	userValidator = auth.NewUserValidator(user, pass)
	jwtValidator = auth.NewJwtValidator()
	totpValidator = auth.NewTOTPValidator(user, totpSecret, otpEnable)
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
