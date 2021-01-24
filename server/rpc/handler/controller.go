package handler

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"google.golang.org/grpc"

	"github.com/cntechpower/anywhere/server/conf"
	pb "github.com/cntechpower/anywhere/server/rpc/definitions"
	"github.com/cntechpower/anywhere/server/server"
	"github.com/cntechpower/anywhere/util"
	"github.com/cntechpower/utils/log"
)

var grpcAddress string

func init() {
	grpcAddress, _ = conf.GetGrpcAddr()
}

func StartRpcServer(s *server.Server, addr string, errChan chan error) {
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

func NewClient(silenceOutput bool) (pb.AnywhereServerClient, error) {
	h := log.NewHeader("grpc")
	cc, err := grpc.Dial(grpcAddress, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{},
			cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			if !silenceOutput {
				log.Infof(h, "calling %v", method)
			}

			err := invoker(ctx, method, req, reply, cc, opts...)
			if !silenceOutput {
				log.Infof(h, "called %v, error: %v", method, err)
			}
			return err
		}))
	if err != nil {
		return nil, err
	}
	return pb.NewAnywhereServerClient(cc), nil
}

func ListAgent() error {
	client, err := NewClient(true)
	if err != nil {
		return err
	}
	res, err := client.ListAgent(context.Background(), &pb.Empty{})
	if err != nil {
		return err
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"UserName", "GroupName", "AgentId", "AgentAddr", "LastAckSend", "LastAckRcv"})
	for _, agent := range res.Agent {
		table.Append([]string{agent.UserName, agent.GroupName, agent.Id, agent.RemoteAddr, agent.LastAckSend, agent.LastAckRcv})
	}
	table.Render()
	return nil
}

func AddProxyConfig(userName, groupName string, remotePort int, localAddr string, isWhiteListOn bool, whiteListIps string) error {

	client, err := NewClient(true)
	if err != nil {
		return err
	}
	input := &pb.AddProxyConfigInput{Config: &pb.ProxyConfig{
		Username:      userName,
		GroupName:     groupName,
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

	client, err := NewClient(true)
	if err != nil {
		return err
	}
	configs, err := client.ListProxyConfigs(context.Background(), &pb.Empty{})
	if err != nil {
		return fmt.Errorf("list proxy config error: %v", err)
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"UserName", "GroupName", "ServerAddr", "LocalAddr", "IsWhiteListOn", "IpWhiteList", "TotalNetFlowsInMB"})
	for _, config := range configs.Config {
		table.Append([]string{
			config.Username, config.GroupName, strconv.Itoa(int(config.RemotePort)), config.LocalAddr,
			util.BoolToString(config.IsWhiteListOn), config.WhiteCidrList,
			strconv.FormatFloat(float64(config.NetworkFlowRemoteToLocalInBytes+config.NetworkFlowLocalToRemoteInBytes)/1024/1024, 'f', 5, 64),
		})
	}
	table.Render()
	return nil
}

func RemoveProxyConfig(userName, groupName string, remotePort int, localAddr string) error {
	client, err := NewClient(true)
	if err != nil {
		return err
	}
	_, err = client.RemoveProxyConfig(context.Background(), &pb.RemoveProxyConfigInput{
		UserName:   userName,
		GroupName:  groupName,
		RemotePort: int64(remotePort),
		LocalAddr:  localAddr,
	})
	return err
}

func LoadProxyConfigFile() error {
	client, err := NewClient(true)
	if err != nil {
		return err
	}
	_, err = client.LoadProxyConfigFile(context.Background(), &pb.Empty{})
	return err
}

func SaveProxyConfigToFile() error {
	client, err := NewClient(true)
	if err != nil {
		return err
	}
	_, err = client.SaveProxyConfigToFile(context.Background(), &pb.Empty{})
	return err
}

func UpdateProxyConfigWhiteList(userName, groupName, localAddr, whiteCidrs string, whiteListEnable bool) error {
	client, err := NewClient(true)
	if err != nil {
		return err
	}
	_, err = client.UpdateProxyConfigWhiteList(context.Background(), &pb.UpdateProxyConfigWhiteListInput{
		UserName:        userName,
		GroupName:       groupName,
		LocalAddr:       localAddr,
		WhiteCidrs:      whiteCidrs,
		WhiteListEnable: whiteListEnable,
	})
	return err
}

func ListConns(groupName string) error {
	client, err := NewClient(true)
	if err != nil {
		return err
	}
	res, err := client.ListConns(context.Background(), &pb.ListConnsInput{
		GroupName: groupName,
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
	table.SetHeader([]string{"UserName", "AgentId", "ConnId", "SrcRemoteAddr", "SrcLocalAddr", "DstRemoteAddr", "DstLocalAddr"})
	for _, conn := range res.Conn {
		table.Append([]string{conn.UserName, conn.AgentId, strconv.Itoa(int(conn.ConnId)), conn.SrcRemoteAddr, conn.SrcLocalAddr, conn.DstRemoteAddr, conn.DstLocalAddr})
	}
	table.Render()
	return nil
}

func KillConn(userName, groupName string, id int) error {
	client, err := NewClient(true)
	if err != nil {
		return err
	}
	_, err = client.KillConnById(context.Background(), &pb.KillConnByIdInput{
		UserName:  userName,
		GroupName: groupName,
		ConnId:    int64(id),
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
	client, err := NewClient(true)
	if err != nil {
		return err
	}
	_, err = client.KillAllConns(context.Background(), &pb.Empty{})
	return err
}

func SendReport() error {
	client, err := NewClient(true)
	if err != nil {
		return err
	}
	_, err = client.SendReport(context.Background(), &pb.Empty{})
	if err != nil {
		return err
	}

	return nil
}
