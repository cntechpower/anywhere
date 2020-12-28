package auth

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/cntechpower/anywhere/log"
	"github.com/cntechpower/anywhere/util"
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
	remotePort int
	agentId    string
	localAddr  string
	logHeader  *log.Header
}

func NewWhiteListValidator(remotePort int, agentId, localAddr, whiteCidrs string, enable bool) (*WhiteListValidator, error) {
	wl, err := util.NewWhiteList(remotePort, agentId, localAddr, whiteCidrs, enable)
	if err != nil {
		return nil, err
	}
	return &WhiteListValidator{
		WhiteList:  wl,
		remotePort: remotePort,
		agentId:    agentId,
		localAddr:  localAddr,
		logHeader:  log.NewHeader("WhiteListValidator"),
	}, nil
}

func (v *WhiteListValidator) AddrInWhiteList(addr string) bool {
	ok := v.WhiteList.AddrInWhiteList(addr)
	labels := []string{strconv.Itoa(v.remotePort), v.agentId, v.localAddr,
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
