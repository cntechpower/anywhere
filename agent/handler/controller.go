package handler

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/cntechpower/utils/log"

	"github.com/cntechpower/anywhere/agent/agent"
	pb "github.com/cntechpower/anywhere/agent/rpc/definitions"
	"github.com/cntechpower/anywhere/util"

	"github.com/olekukonko/tablewriter"

	"google.golang.org/grpc"
)

func StartRpcServer(agent *agent.Agent, addr string, errChan chan error) {
	if err := util.CheckAddrValid(addr); err != nil {
		errChan <- err
		return
	}
	// The agent admin API has no authentication, so it must never be exposed
	// beyond the loopback interface.
	if !util.IsLoopbackAddr(addr) {
		errChan <- fmt.Errorf("grpc address %s is not loopback; refusing to expose an unauthenticated admin API", addr)
		return
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		errChan <- err
		return
	}
	grpcServer := grpc.NewServer()
	pb.RegisterAnywhereAgentServer(grpcServer, &anywhereAgentRpcHandler{a: agent, logHeader: log.NewHeader("agent")})
	if err := grpcServer.Serve(l); err != nil {
		errChan <- err
	}
}

func NewClient(addr string) (pb.AnywhereAgentClient, error) {
	cc, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return pb.NewAnywhereAgentClient(cc), nil
}

func ListConnections(grpcAddr string) error {
	client, err := NewClient(grpcAddr)
	if err != nil {
		return err
	}
	res, err := client.ListConnections(context.Background(), &pb.Empty{})
	if err != nil {
		return err
	}
	if len(res.Conn) == 0 {
		fmt.Println("no conn exist")
		return nil
	}
	table := tablewriter.NewTable(os.Stdout)
	table.Header("ConnId", "SrcRemoteAddr", "SrcLocalAddr", "DstRemoteAddr", "DstLocalAddr")
	for _, conn := range res.Conn {
		table.Append(strconv.Itoa(int(conn.ConnId)), conn.SrcRemoteAddr, conn.SrcLocalAddr, conn.DstRemoteAddr, conn.DstLocalAddr)
	}
	_ = table.Render()
	return nil
}

func KillConn(grpcAddr string, id int) error {
	client, err := NewClient(grpcAddr)
	if err != nil {
		return err
	}
	_, err = client.KillConnById(context.Background(), &pb.KillConnByIdInput{
		ConnId: int64(id),
	})
	return err
}

func FlushConns(grpcAddr string) error {
	fmt.Println("ATTENTION: are you sure to flush all connections?")
	fmt.Println("y/n ?")
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	if strings.TrimSpace(text) != "y" {
		fmt.Println("cancelled")
		return nil
	}
	client, err := NewClient(grpcAddr)
	if err != nil {
		return err
	}
	_, err = client.KillAllConnections(context.Background(), &pb.Empty{})
	return err
}

func ShowStatus(grpcAddr string) error {
	client, err := NewClient(grpcAddr)
	if err != nil {
		return err
	}
	res, err := client.ShowStatus(context.Background(), &pb.Empty{})
	if err != nil {
		return err
	}

	table := tablewriter.NewTable(os.Stdout)
	table.Header("AgentId", "LocalAddr", "ServerAddr", "LastAckSend", "LastAckRcv")
	table.Append(res.AgentId, res.LocalAddr, res.ServerAddr, res.LastAckSendTime, res.LastAckRcvTime)
	_ = table.Render()
	return nil
}
