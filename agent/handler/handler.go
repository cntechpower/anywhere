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
	pb.UnimplementedAnywhereAgentServer
}

func (h *anywhereAgentRpcHandler) ListConnections(ctx context.Context, empty *pb.Empty) (res *pb.Conns, err error) {
	connections, err := h.a.ListJoinedConn()
	if err != nil {
		return
	}
	res = &pb.Conns{
		Conn: make([]*pb.Conn, 0),
	}

	for _, conn := range connections {
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

func (h *anywhereAgentRpcHandler) KillConnById(_ context.Context, input *pb.KillConnByIdInput) (*pb.Empty, error) {
	return &pb.Empty{}, h.a.KillJoinedConnById(uint(input.ConnId))
}

func (h *anywhereAgentRpcHandler) KillAllConnections(_ context.Context, _ *pb.Empty) (*pb.Empty, error) {
	h.a.FlushJoinedConn()
	return &pb.Empty{}, nil
}

func (h *anywhereAgentRpcHandler) ShowStatus(_ context.Context, _ *pb.Empty) (*pb.ShowStatusOutput, error) {
	s := h.a.GetStatus()
	return &pb.ShowStatusOutput{
		AgentId:         s.Id,
		LocalAddr:       s.LocalAddr,
		ServerAddr:      s.ServerAddr,
		LastAckSendTime: s.LastAckSend,
		LastAckRcvTime:  s.LastAckRcv,
	}, nil
}
