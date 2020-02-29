package frontEnd

import (
	"anywhere/log"
	"anywhere/util"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Start(addr string, closeChan chan struct{}) {
	if err := util.CheckAddrValid(addr); err != nil {
		panic(err)
	}
	if !util.CheckPathExist("./static") {
		panic("static dir not found")
	}
	router := gin.Default()
	router.StaticFS("/react/", http.Dir("./static"))

	//try to render template to support front end router
	//router.StaticFS("/react/static/", http.Dir("./static/static"))
	//router.LoadHTMLFiles("./static/index.html")
	//router.Any("/react/", func(c *gin.Context) {
	//	c.HTML(http.StatusOK, "index.html", nil)
	//})
	l := log.GetCustomLogger("web_interface")
	l.Infof("start as %s", addr)
	go func() {
		if err := router.Run(addr); err != nil {
			l.Errorf("%s", err)
		}
	}()
	select {
	case <-closeChan:
		l.Infof("closing")
		return
	}
}
