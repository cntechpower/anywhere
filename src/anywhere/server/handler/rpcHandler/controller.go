package rpcHandler

import (
	"anywhere/log"
	"anywhere/server/anywhereServer"
	pb "anywhere/server/rpc/definitions"
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

var grpcAddress string

func init() {
	grpcAddress, _ = anywhereServer.GetGrpcAddr()
}

func StartRpcServer(s *anywhereServer.Server, addr string, errChan chan error) {
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
	pb.RegisterAnywhereServerServer(grpcServer, GetRpcHandlers(s))
	if err := grpcServer.Serve(l); err != nil {
		errChan <- err
	}
}

func NewClient() (pb.AnywhereServerClient, error) {
	h := log.NewHeader("grpc")
	cc, err := grpc.Dial(grpcAddress, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{},
			cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			log.Infof(h, "calling %v", method)
			defer log.Infof(h, "called %v", method)
			return invoker(ctx, method, req, reply, cc, opts...)
		}))
	if err != nil {
		return nil, err
	}
	return pb.NewAnywhereServerClient(cc), nil
}

func ListAgent() error {
	client, err := NewClient()
	if err != nil {
		return err
	}
	res, err := client.ListAgent(context.Background(), &pb.Empty{})
	if err != nil {
		return err
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"AgentId", "AgentAddr", "LastAckSend", "LastAckRcv"})
	for _, agent := range res.Agent {
		table.Append([]string{agent.AgentId, agent.AgentRemoteAddr, agent.AgentLastAckSend, agent.AgentLastAckRcv})
	}
	table.Render()
	return nil
}

func AddProxyConfig(agentId string, remotePort int, localAddr string, isWhiteListOn bool, whiteListIps string) error {

	client, err := NewClient()
	if err != nil {
		return err
	}
	input := &pb.AddProxyConfigInput{Config: &pb.ProxyConfig{
		AgentId:       agentId,
		RemotePort:    int64(remotePort),
		LocalAddr:     localAddr,
		IsWhiteListOn: isWhiteListOn,
		WhiteCidrList: whiteListIps,
	}}
	if _, err = client.AddProxyConfig(context.Background(), input); err != nil {
		return fmt.Errorf("add proxy config error: %v", err)
	}
	return nil

}

func ListProxyConfigs() error {
	boolToString := func(b bool) string {
		if b {
			return "ON"
		}
		return "OFF"
	}
	client, err := NewClient()
	if err != nil {
		return err
	}
	configs, err := client.ListProxyConfigs(context.Background(), &pb.Empty{})
	if err != nil {
		return fmt.Errorf("list proxy config error: %v", err)
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"AgentId", "ServerAddr", "LocalAddr", "IsWhiteListOn", "IpWhiteList"})
	for _, config := range configs.Config {
		table.Append([]string{config.AgentId, strconv.Itoa(int(config.RemotePort)), config.LocalAddr, boolToString(config.IsWhiteListOn), config.WhiteCidrList})
	}
	table.Render()
	return nil
}

func RemoveProxyConfig(agentId, localAddr string) error {
	client, err := NewClient()
	if err != nil {
		return err
	}
	_, err = client.RemoveProxyConfig(context.Background(), &pb.RemoveProxyConfigInput{
		AgentId:   agentId,
		LocalAddr: localAddr,
	})
	return err
}

func LoadProxyConfigFile() error {
	client, err := NewClient()
	if err != nil {
		return err
	}
	_, err = client.LoadProxyConfigFile(context.Background(), &pb.Empty{})
	return err
}

func SaveProxyConfigToFile() error {
	client, err := NewClient()
	if err != nil {
		return err
	}
	_, err = client.SaveProxyConfigToFile(context.Background(), &pb.Empty{})
	return err
}

func UpdateProxyConfigWhiteList(agentId, localAddr, whiteCidrs string, whiteListEnable bool) error {
	client, err := NewClient()
	if err != nil {
		return err
	}
	_, err = client.UpdateProxyConfigWhiteList(context.Background(), &pb.UpdateProxyConfigWhiteListInput{
		AgentId:         agentId,
		LocalAddr:       localAddr,
		WhiteCidrs:      whiteCidrs,
		WhiteListEnable: whiteListEnable,
	})
	return err
}

func ListConns(agentId string) error {
	client, err := NewClient()
	if err != nil {
		return err
	}
	res, err := client.ListConns(context.Background(), &pb.ListConnsInput{
		AgentId: agentId,
	})
	if err != nil {
		return err
	}
	if len(res.Conn) == 0 {
		fmt.Println("no conn exist")
		return nil
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"AgentId", "ConnId", "SrcRemoteAddr", "SrcLocalAddr", "DstRemoteAddr", "DstLocalAddr"})
	for _, conn := range res.Conn {
		table.Append([]string{conn.AgentId, strconv.Itoa(int(conn.ConnId)), conn.SrcRemoteAddr, conn.SrcLocalAddr, conn.DstRemoteAddr, conn.DstLocalAddr})
	}
	table.Render()
	return nil
}

func KillConn(agentId string, id int) error {
	client, err := NewClient()
	if err != nil {
		return err
	}
	_, err = client.KillConnById(context.Background(), &pb.KillConnByIdInput{
		AgentId: agentId,
		ConnId:  int64(id),
	})
	return err
}

func FlushConns() error {
	fmt.Println("ATTENTION: are you sure to flush all connections?")
	fmt.Println("y/n ?")
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	if strings.TrimSpace(text) != "y" {
		fmt.Println("cancelled")
		return nil
	}
	client, err := NewClient()
	if err != nil {
		return err
	}
	_, err = client.KillAllConns(context.Background(), &pb.Empty{})
	return err
}
