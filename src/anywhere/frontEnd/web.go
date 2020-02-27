package frontEnd

import (
	"anywhere/log"
	"anywhere/util"
	"net/http"
)

func Start(addr string, closeChan chan struct{}) {
	if err := util.CheckAddrValid(addr); err != nil {
		panic(err)
	}
	if !util.CheckPathExist("./static") {
		panic("static dir not found")
	}
	//http.Handle("/", http.FileServer(http.Dir("static")))
	http.Handle("/react/", http.StripPrefix("/react/", http.FileServer(http.Dir("./static"))))
	l := log.GetCustomLogger("web_interface")
	l.Infof("start as %s", addr)
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			l.Errorf("%s", err)
		}
	}()
	select {
	case <-closeChan:
		l.Infof("closing")
		return
	}
}
