package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cntechpower/anywhere/server/conf"

	mhttp "github.com/cntechpower/utils/monitor/http"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"github.com/cntechpower/anywhere/util"
)

var (
	RespUserPassIsRequired = gin.H{
		"code": http.StatusBadRequest,
		"data": "用户名和密码不能为空",
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

const (
	sessionMaxAge = 7 * 24 * 3600
	shutdownGrace = 5 * time.Second
)

// addUIRouter serves the built SPA from dir and falls back to index.html for
// client side routes. The previous implementation depended on the unmaintained
// gin-gonic/contrib/static package.
func addUIRouter(router *gin.Engine, dir string) error {
	if !util.CheckPathExist(dir) {
		return fmt.Errorf("static dir %q not found", dir)
	}
	fileServer := http.FileServer(http.Dir(dir))
	indexFile := filepath.Join(dir, "index.html")

	router.NoRoute(func(c *gin.Context) {
		requestPath := c.Request.URL.Path
		if strings.HasPrefix(requestPath, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "data": "api not found"})
			return
		}
		cleanPath := filepath.Clean("/" + requestPath)
		target := filepath.Join(dir, cleanPath)
		if info, err := os.Stat(target); err == nil && !info.IsDir() {
			if strings.HasPrefix(cleanPath, "/assets/") {
				// Vite emits content hashed file names, so they can be cached
				// forever.
				c.Header("Cache-Control", "public, max-age=31536000, immutable")
			}
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}
		if filepath.Ext(cleanPath) != "" {
			c.Status(http.StatusNotFound)
			return
		}
		c.File(indexFile)
	})
	return nil
}

func addAPIRouter(router *gin.Engine, restHandler http.Handler) {
	apiRouter := router.Group("/api")
	apiRouter.Any("/*any", gin.WrapH(restHandler))
}

// securityHeaders sets a few defensive headers on every response.
func securityHeaders(c *gin.Context) {
	h := c.Writer.Header()
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("X-Frame-Options", "DENY")
	h.Set("Referrer-Policy", "same-origin")
	c.Next()
}

// csrfProtect rejects state changing requests whose Origin/Referer points to a
// different host. Combined with SameSite=Lax cookies this closes CSRF on the
// cookie authenticated API.
func csrfProtect(c *gin.Context) {
	switch c.Request.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		c.Next()
		return
	}
	origin := c.GetHeader("Origin")
	if origin == "" {
		origin = c.GetHeader("Referer")
	}
	if origin != "" {
		parsed, err := url.Parse(origin)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code": http.StatusForbidden,
				"data": "invalid origin header",
			})
			return
		}
		if parsed.Host != c.Request.Host {
			// The Vite dev server runs on another port (same site, different
			// origin) while CORS is enabled, so allow it explicitly.
			if conf.Conf.UiConfig.DebugMode && (parsed.Host == "localhost:3000" || parsed.Host == "127.0.0.1:3000") {
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code": http.StatusForbidden,
				"data": "cross origin request rejected",
			})
			return
		}
	}
	c.Next()
}

func (a *API) StartUIAndAPIService(ctx context.Context, restHandler http.Handler, errChan chan error) {
	if !conf.Conf.UiConfig.IsWebEnable {
		return
	}
	if err := util.CheckAddrValid(conf.Conf.UiConfig.WebAddr); err != nil {
		errChan <- err
		return
	}
	// Set the gin mode before gin.New() so the startup banner is not printed in
	// release mode.
	if conf.Conf.UiConfig.DebugMode {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()

	options := []mhttp.GinMiddlewareOption{
		mhttp.WithLog(true, true),
		mhttp.WithBlackList([]string{"/favicon.ico"}),
	}
	if conf.Conf.UiConfig.DebugMode {
		options = append(options, mhttp.WithTrace())
	}
	router.Use(mhttp.GinMiddleware(options...))
	router.Use(securityHeaders)
	router.Use(csrfProtect)
	router.Use(gin.Recovery())

	if conf.Conf.UiConfig.DebugMode {
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
			SkipPaths: []string{"/assets"},
		}))
		router.Use(cors.New(cors.Config{
			AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}))
	}

	store := cookie.NewStore(conf.SessionKey())
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   sessionMaxAge,
		HttpOnly: true,
		Secure:   conf.Conf.UiConfig.EnableTLS,
		SameSite: http.SameSiteLaxMode,
	})
	router.Use(sessions.Sessions("anywhere", store))

	if conf.Conf.UiConfig.SkipLogin {
		log.Print("WARNING: ui_config.skip_login is enabled, the management API is NOT authenticated. " +
			"Never enable this on a public server.")
	} else {
		router.Use(a.SessionFilter)
	}

	router.POST("/user_login", a.UserLogin)
	router.POST("/user_logout", a.UserLogout)
	if err := addUIRouter(router, "./static"); err != nil {
		// API-only deployments are valid; the UI just stays unavailable.
		log.Printf("WARNING: %v, the web UI is disabled. Run `make ui` to build it.", err)
	}
	addAPIRouter(router, restHandler)

	srv := &http.Server{
		Addr:              conf.Conf.UiConfig.WebAddr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownGrace)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	if conf.Conf.UiConfig.EnableTLS {
		if conf.Conf.HttpSSL == nil {
			errChan <- fmt.Errorf("ui_config.enable_tls is true but http_ssl_config is missing")
			return
		}
		errChan <- srv.ListenAndServeTLS(conf.Conf.HttpSSL.CertFile, conf.Conf.HttpSSL.KeyFile)
	} else {
		errChan <- srv.ListenAndServe()
	}
}
