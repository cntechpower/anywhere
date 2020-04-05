package handler

import (
	"anywhere/agent/anywhereAgent"
	pb "anywhere/agent/rpc/definitions"
	"anywhere/log"
	"anywhere/util"
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"

	"google.golang.org/grpc"
)

func StartRpcServer(agent *anywhereAgent.Agent, addr string, errChan chan error) {
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
	pb.RegisterAnywhereServer(grpcServer, &anywhereAgentRpcHandler{a: agent, l: log.GetCustomLogger("grpc_handler")})
	if err := grpcServer.Serve(l); err != nil {
		errChan <- err
	}
}

func NewClient(addr string) (pb.AnywhereClient, error) {
	cc, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return pb.NewAnywhereClient(cc), nil
}

func ListConns(grpcAddr string) error {
	client, err := NewClient(grpcAddr)
	if err != nil {
		return err
	}
	res, err := client.ListConns(context.Background(), &pb.Empty{})
	if err != nil {
		return err
	}
	if len(res.Conn) == 0 {
		fmt.Println("no conn exist")
		return nil
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"ConnId", "SrcRemoteAddr", "SrcLocalAddr", "DstRemoteAddr", "DstLocalAddr"})
	for _, conn := range res.Conn {
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
	_, err = client.KillAllConns(context.Background(), &pb.Empty{})
	return err
}
