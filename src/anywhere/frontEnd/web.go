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

	http.Handle("/", http.FileServer(http.Dir("static")))
	l := log.GetCustomLogger("web_interface")
	l.Infof("start as %s", addr)
	go http.ListenAndServe(addr, nil)
	select {
	case <-closeChan:
		l.Infof("closing")
		return
	}
}
