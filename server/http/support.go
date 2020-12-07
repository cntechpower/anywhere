package http

import (
	"net"

	v1 "github.com/cntechpower/anywhere/server/restapi/api/restapi/operations"
)

func GetV1SupportIP(params v1.GetV1SupportIPParams) (string, error) {
	addr, err := net.ResolveTCPAddr("tcp", params.HTTPRequest.RemoteAddr)
	if err != nil {
		return "", err
	}
	return addr.IP.String(), nil
}
