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
	pb "github.com/cntechpower/anywhere/gen/go/github.com/cntechpower/anywhere/gen/go/agent_pb"
	"github.com/cntechpower/anywhere/util"

	"github.com/olekukonko/tablewriter"

	"google.golang.org/grpc"
)

func StartRpcServer(agent *agent.Agent, addr string, errChan chan error) {
	if err := util.CheckAddrValid(addr); err != nil {
		errChan <- err
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
	if len(res.Conns) == 0 {
		fmt.Println("no conn exist")
		return nil
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"ConnId", "SrcRemoteAddr", "SrcLocalAddr", "DstRemoteAddr", "DstLocalAddr"})
	for _, conn := range res.Conns {
		table.Append([]string{strconv.Itoa(int(conn.ConnId)), conn.SrcRemoteAddr, conn.SrcLocalAddr, conn.DstRemoteAddr, conn.DstLocalAddr})
	}
	table.Render()
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

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"AgentId", "LocalAddr", "ServerAddr", "LastAckSend", "LastAckRcv"})
	table.Append([]string{res.AgentId, res.LocalAddr, res.ServerAddr, res.LastAckSendTime, res.LastAckRcvTime})
	table.Render()
	return nil
}
