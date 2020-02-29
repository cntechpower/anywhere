package main

import (
	"anywhere/server/restapi/api/restapi"
	"anywhere/server/restapi/api/restapi/operations"
	"anywhere/util"
	"fmt"
	"net/http"

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

	router.LoadHTMLFiles("./static/index.html")
	renderIndex := func(c *gin.Context) {
		user, ok := c.Get(gin.AuthUserKey)
		if !ok {
			c.JSON(http.StatusOK, gin.H{"user": "no user", "message": "Forbidden"})
		}
		if _, ok := secrets[user.(string)]; ok {
			c.HTML(http.StatusOK, "index.html", nil)
		} else {
			c.JSON(http.StatusForbidden, gin.H{"user": user, "message": "Forbidden"})
		}

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

func startUIAndAPIService(addr, certFile, keyFile string, errChan chan error) {
	if err := util.CheckAddrValid(addr); err != nil {
		errChan <- err
	}
	router := gin.New()
	//store := cookie.NewStore([]byte("secret"))
	router.Use(gin.BasicAuth(gin.Accounts{"dujinyang": "suya9495"}))
	//router.Use(sessions.Sessions("mysession", store))
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
