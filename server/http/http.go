package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/cntechpower/anywhere/server/auth"
	"github.com/cntechpower/anywhere/server/server"
	"github.com/cntechpower/anywhere/util"
	"github.com/cntechpower/utils/log"
)

var userValidator *auth.UserValidator
var whiteListValidator *auth.WhiteListValidator
var jwtValidator *auth.JwtValidator
var serverInst *server.Server

var (
	RespUserPassIsRequired = gin.H{
		"code": http.StatusBadRequest,
		"data": "用户名/密码/动态码不能为空",
	}
	RespUserPassWrong = gin.H{
		"code": http.StatusUnauthorized,
		"data": "用户名或密码错误",
	}

	RespUserLoginSuccess = gin.H{
		"code": http.StatusOK,
		"data": "登陆成功",
	}
)

func addUIRouter(router *gin.Engine) error {
	if !util.CheckPathExist("./static") {
		return fmt.Errorf("static dir not found")
	}
	staticHandler := static.Serve("/", static.LocalFile("./static", true))
	router.Use(staticHandler)
	router.NoRoute(func(c *gin.Context) {
		c.File("./static/index.html")
	})

	return nil
}

func addAPIRouter(router *gin.Engine, restHandler http.Handler) error {
	apiRouter := router.Group("/api")
	apiRouter.Any("/*any", gin.WrapH(restHandler))
	return nil
}

func StartUIAndAPIService(restHandler http.Handler, serverI *server.Server, addr string, errChan chan error, skipLogin, debug bool, reportWhiteCidrs string) {
	if err := util.CheckAddrValid(addr); err != nil {
		errChan <- err
	}
	serverInst = serverI
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
			SkipPaths: []string{"/static"},
		}))
		router.Use(cors.New(cors.Config{
			AllowAllOrigins:        true,
			AllowWildcard:          true,
			AllowBrowserExtensions: true,
			AllowWebSockets:        true,
			AllowFiles:             true,
		}))
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	userValidator = server.GetServerInstance().GetUserValidator()
	jwtValidator = auth.NewJwtValidator()
	//session auth
	store := cookie.NewStore([]byte(util.RandString(16)))
	router.Use(sessions.Sessions("anywhere", store))
	//support frontend development
	if !skipLogin {
		router.Use(sessionFilter)
	}

	router.POST("/user_login", userLogin)
	if err := addUIRouter(router); err != nil {
		errChan <- err
	}
	if err := addAPIRouter(router, restHandler); err != nil {
		errChan <- err
	}
	_, port, _ := util.GetIpPortByAddr(addr)
	var err error
	whiteListValidator, err = auth.NewWhiteListValidator(port, "frontend", addr, reportWhiteCidrs, true)
	if err != nil {
		panic(err)
	}
	router.GET("/report", whiteListValidator.GinHandler, func(ctx *gin.Context) {
		html, err := serverInst.GetHtmlReport(log.NewHeader("web"))
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.Header("Content-Type", "text/html; charset=utf-8")
		ctx.String(http.StatusOK, html)
	})
	router.GET("/metrics", whiteListValidator.GinHandler, gin.WrapH(promhttp.Handler()))
	errChan <- router.Run(addr)

}
