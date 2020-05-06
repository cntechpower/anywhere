package handler

import (
	"anywhere/agent/anywhereAgent"
	pb "anywhere/agent/rpc/definitions"
	"anywhere/log"

	"context"
)

type anywhereAgentRpcHandler struct {
	a         *anywhereAgent.Agent
	logHeader *log.Header
}

func (h *anywhereAgentRpcHandler) ListConns(ctx context.Context, empty *pb.Empty) (*pb.Conns, error) {
	log.Infof(h.logHeader, "calling list conns")
	defer log.Infof(h.logHeader, "called list conns")
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
	log.Infof(h.logHeader, "calling kill conn %v", input.ConnId)
	defer log.Infof(h.logHeader, "called kill conn %v", input.ConnId)
	return &pb.Empty{}, h.a.KillJoinedConnById(int(input.ConnId))
}

func (h *anywhereAgentRpcHandler) KillAllConns(ctx context.Context, empty *pb.Empty) (*pb.Empty, error) {
	log.Infof(h.logHeader, "calling flush conns")
	defer log.Infof(h.logHeader, "called flush conns")
	h.a.FlushJoinedConns()
	return &pb.Empty{}, nil
}

func (h *anywhereAgentRpcHandler) ShowStatus(ctx context.Context, empty *pb.Empty) (*pb.ShowStatusOutput, error) {
	log.Infof(h.logHeader, "calling show status")
	defer log.Infof(h.logHeader, "called show status")
	s := h.a.GetStatus()
	return &pb.ShowStatusOutput{
		AgentId:         s.Id,
		LocalAddr:       s.LocalAddr,
		ServerAddr:      s.ServerAddr,
		LastAckSendTime: s.LastAckSend,
		LastAckRcvTime:  s.LastAckRcv,
	}, nil
}
