package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cntechpower/anywhere/util"
	"github.com/cntechpower/utils/log"
	"github.com/gin-gonic/gin"
)

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

func (v *WhiteListValidator) AddrInWhiteList(ctx context.Context, addr string) bool {
	ok := v.WhiteList.AddrInWhiteList(ctx, addr)
	return ok
}

func (v *WhiteListValidator) GinHandler(ctx *gin.Context) {
	if !v.AddrInWhiteList(ctx, ctx.Request.RemoteAddr) {
		_ = ctx.AbortWithError(http.StatusUnauthorized, fmt.Errorf("your ip address is not in white list"))
	}
}
