package handler

import (
	"anywhere/agent/anywhereAgent"
	pb "anywhere/agent/rpc/definitions"

	"github.com/sirupsen/logrus"

	"context"
)

type anywhereAgentRpcHandler struct {
	a *anywhereAgent.Agent
	l *logrus.Entry
}

func (h *anywhereAgentRpcHandler) ListConns(ctx context.Context, empty *pb.Empty) (*pb.Conns, error) {
	h.l.Infof("calling list conns")
	defer h.l.Infof("called list conns")
	conns := h.a.ListJoinedConns()
	res := &pb.Conns{
		Conn: make([]*pb.Conn, 0),
	}

	for _, conn := range conns {
		res.Conn = append(res.Conn, &pb.Conn{
			ConnId:        int64(conn.ConnId),
			SrcRemoteAddr: conn.SrcRemoteAddr,
			SrcLocalAddr:  conn.SrcLocalAddr,
			DstRemoteAddr: conn.DstRemoteAddr,
			DstLocalAddr:  conn.DstLocalAddr,
		})
	}

	return res, nil
}

func (h *anywhereAgentRpcHandler) KillConnById(ctx context.Context, input *pb.KillConnByIdInput) (*pb.Empty, error) {
	h.l.Infof("calling kill conn %v", input.ConnId)
	defer h.l.Infof("called kill conn %v", input.ConnId)
	return &pb.Empty{}, h.a.KillJoinedConnById(int(input.ConnId))
}

func (h *anywhereAgentRpcHandler) KillAllConns(ctx context.Context, empty *pb.Empty) (*pb.Empty, error) {
	h.l.Infof("calling flush conns")
	defer h.l.Infof("called flush conns")
	h.a.FlushJoinedConns()
	return &pb.Empty{}, nil
}
