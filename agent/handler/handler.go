package handler

import (
	"github.com/cntechpower/utils/log"

	"github.com/cntechpower/anywhere/agent/agent"
	pb "github.com/cntechpower/anywhere/agent/rpc/definitions"

	"context"
)

type anywhereAgentRpcHandler struct {
	a         *agent.Agent
	logHeader *log.Header
}

func (h *anywhereAgentRpcHandler) ListConns(ctx context.Context, empty *pb.Empty) (res *pb.Conns, err error) {
	conns, err := h.a.ListJoinedConn()
	if err != nil {
		return
	}
	res = &pb.Conns{
		Conn: make([]*pb.Conn, 0),
	}

	for _, conn := range conns {
		res.Conn = append(res.Conn, &pb.Conn{
			ConnId:        int64(conn.ID),
			SrcRemoteAddr: conn.SrcRemoteAddr,
			SrcLocalAddr:  conn.SrcLocalAddr,
			DstRemoteAddr: conn.DstRemoteAddr,
			DstLocalAddr:  conn.DstLocalAddr,
		})
	}

	return res, nil
}

func (h *anywhereAgentRpcHandler) KillConnById(ctx context.Context, input *pb.KillConnByIdInput) (*pb.Empty, error) {
	return &pb.Empty{}, h.a.KillJoinedConnById(uint(input.ConnId))
}

func (h *anywhereAgentRpcHandler) KillAllConns(ctx context.Context, empty *pb.Empty) (*pb.Empty, error) {
	h.a.FlushJoinedConn()
	return &pb.Empty{}, nil
}

func (h *anywhereAgentRpcHandler) ShowStatus(ctx context.Context, empty *pb.Empty) (*pb.ShowStatusOutput, error) {
	s := h.a.GetStatus()
	return &pb.ShowStatusOutput{
		AgentId:         s.Id,
		LocalAddr:       s.LocalAddr,
		ServerAddr:      s.ServerAddr,
		LastAckSendTime: s.LastAckSend,
		LastAckRcvTime:  s.LastAckRcv,
	}, nil
}
