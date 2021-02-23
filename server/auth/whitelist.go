package auth

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/cntechpower/anywhere/util"
	"github.com/cntechpower/utils/log"
)

var whiteListDenyCount = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "whitelist_deny_count"},
	[]string{"remote_port", "agent_id", "local_addr", "ip"})

var whiteListOkCount = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "whitelist_ok_count"},
	[]string{"remote_port", "agent_id", "local_addr", "ip"})

func init() {
	prometheus.MustRegister(whiteListDenyCount)
	prometheus.MustRegister(whiteListOkCount)
}

type WhiteListValidator struct {
	*util.WhiteList
	logHeader *log.Header
	//port, name, localAddr are used for prometheus metrics lables
	port      int
	name      string
	localAddr string
}

func NewWhiteListValidator(port int, name, localAddr, whiteCidrs string, enable bool) (*WhiteListValidator, error) {
	wl, err := util.NewWhiteList(port, name, localAddr, whiteCidrs, enable)
	if err != nil {
		return nil, err
	}
	return &WhiteListValidator{
		WhiteList: wl,
		port:      port,
		name:      name,
		localAddr: localAddr,
		logHeader: log.NewHeader("WhiteListValidator"),
	}, nil
}

func (v *WhiteListValidator) AddrInWhiteList(addr string) bool {
	ok := v.WhiteList.AddrInWhiteList(addr)
	labels := []string{strconv.Itoa(v.port), v.name, v.localAddr,
		strings.Split(addr, ":")[0]}
	if !ok {
		whiteListDenyCount.WithLabelValues(labels...).Add(1)
	} else {
		whiteListOkCount.WithLabelValues(labels...).Add(1)
	}
	return ok
}

func (v *WhiteListValidator) GinHandler(ctx *gin.Context) {
	if !v.AddrInWhiteList(ctx.Request.RemoteAddr) {
		_ = ctx.AbortWithError(http.StatusUnauthorized, fmt.Errorf("your ip address is not in white list"))
	}
}
