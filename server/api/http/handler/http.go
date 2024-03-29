package handler

import (
	"fmt"
	"net/http"
	xos "os"
	"time"

	"github.com/cntechpower/anywhere/server/conf"

	"github.com/cntechpower/anywhere/server/api/auth"
	"github.com/cntechpower/anywhere/server/server"
	"github.com/cntechpower/anywhere/util"
	mhttp "github.com/cntechpower/utils/monitor/http"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
)

var userValidator *auth.UserValidator
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

//nolint:funlen
func StartUIAndAPIService(restHandler http.Handler, serverI *server.Server, errChan chan error) {
	if !conf.Conf.UiConfig.IsWebEnable {
		return
	}
	if err := util.CheckAddrValid(conf.Conf.UiConfig.WebAddr); err != nil {
		errChan <- err
	}
	serverInst = serverI
	router := gin.New()

	options := []mhttp.GinMiddlewareOption{
		mhttp.WithLog(true, true),
		mhttp.WithBlackList([]string{"/favicon.ico"}),
	}
	traceAddr := xos.Getenv("TRACE_ADDR")
	if traceAddr != "" {
		options = append(options, mhttp.WithTrace())
	}
	router.Use(mhttp.GinMiddleware(options...))

	if conf.Conf.UiConfig.DebugMode {
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
	if !conf.Conf.UiConfig.SkipLogin {
		router.Use(SessionFilter)
	}

	router.POST("/user_login", UserLogin)
	if err := addUIRouter(router); err != nil {
		errChan <- err
	}
	if err := addAPIRouter(router, restHandler); err != nil {
		errChan <- err
	}
	errChan <- router.RunTLS(conf.Conf.UiConfig.WebAddr, conf.Conf.HttpSSL.CertFile, conf.Conf.HttpSSL.KeyFile)
}
