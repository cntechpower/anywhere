package rpcServer

import (
	"anywhere/server/handler"
	"anywhere/util"
	"context"
	"fmt"
	"net"
	"os"

	"github.com/olekukonko/tablewriter"

	pb "anywhere/server/rpc/definitions"

	"google.golang.org/grpc"
)

func StartRpcServer(port int, errChan chan error) {
	addr, err := util.GetAddrByIpPort("127.0.0.1", port)
	if err != nil {
		errChan <- err
		return
	}

	l, err := net.Listen("tcp", addr.String())
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

func rpcCall(port int, f func(cc pb.AnywhereServerClient) error) error {
	addr, err := util.GetAddrByIpPort("127.0.0.1", port)
	if err != nil {
		return err
	}
	cc, err := grpc.Dial(addr.String(), grpc.WithInsecure())
	if err != nil {
		return err
	}
	c := pb.NewAnywhereServerClient(cc)
	if err := f(c); err != nil {
		return err
	}
	return nil
}

func RpcListAgent(port int) error {
	var res *pb.Agents
	var err error
	f := func(client pb.AnywhereServerClient) error {
		res, err = client.ListAgent(context.Background(), &pb.Empty{})
		return err
	}
	if err := rpcCall(port, f); err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"AgentId", "AgentAddr", "LastAck", "Status"})
	for _, agent := range res.Agent {
		table.Append([]string{agent.AgentId, agent.AgentVersion, agent.AgentRemoteAddr})
	}
	table.Render()

	return nil
}

func RpcAddProxyConfig(port int, agentId, remotePort, localIp, localPort string) error {
	f := func(client pb.AnywhereServerClient) error {
		input := &pb.AddProxyConfigInput{
			AgentId:    agentId,
			RemotePort: remotePort,
			LocalIp:    localIp,
			LocalPort:  localPort,
		}
		_, err := client.AddProxyConfig(context.Background(), input)
		if err != nil {
			return fmt.Errorf("add proxy config error: %v", err)
		}
		return nil
	}
	return rpcCall(port, f)
}
