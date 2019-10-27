package rpcServer

import (
	"anywhere/server/handler"
	"anywhere/util"
	"crypto/tls"

	pb "anywhere/server/rpc/definitions"

	"google.golang.org/grpc"
)

func StartRpcServer(port int, config *tls.Config, errChan chan error) {
	addr, err := util.GetAddrByIpPort("0.0.0.0", port)
	if err != nil {
		errChan <- err
		return
	}

	l, err := tls.Listen("tcp", addr.String(), config)
	if err != nil {
		errChan <- err
		return
	}
	grpcServer := grpc.NewServer()
	pb.RegisterAnywhereServerServer(grpcServer, handler.GetRpcHandlers())
	if err := grpcServer.Serve(l); err != nil {
		errChan <- err
	}
}
