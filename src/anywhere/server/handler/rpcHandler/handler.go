package rpcHandler

import (
	pb "anywhere/server/rpc/definitions"
	"anywhere/util"
	"context"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"google.golang.org/grpc"
)
import "anywhere/server/anywhereServer"

var grpcAddress *net.TCPAddr

type rpcHandlers struct {
}

func GetRpcHandlers() *rpcHandlers {
	return &rpcHandlers{}
}

func (h *rpcHandlers) ListAgent(ctx context.Context, empty *pb.Empty) (*pb.Agents, error) {
	s := anywhereServer.GetServerInstance()
	if s == nil {
		return nil, fmt.Errorf("anywhere server not init")
	}
	res := &pb.Agents{
		Agent: make([]*pb.Agent, 0),
	}
	agents := s.ListAgentInfo()
	for _, agent := range agents {
		res.Agent = append(res.Agent, &pb.Agent{
			AgentId:          agent.Id,
			AgentRemoteAddr:  agent.RemoteAddr,
			AgentLastAckRcv:  agent.LastAckRcv,
			AgentLastAckSend: agent.LastAckSend,
		})
	}
	return res, nil
}

func (h *rpcHandlers) AddProxyConfig(ctx context.Context, input *pb.AddProxyConfigInput) (*pb.Empty, error) {
	if input.Config == nil {
		return nil, fmt.Errorf("config not vaild: nil")
	}
	s := anywhereServer.GetServerInstance()
	if s == nil {
		return nil, fmt.Errorf("anywhere server not init")
	}
	config := input.Config
	remotePort, err := strconv.Atoi(config.RemotePort)
	if err != nil {
		return nil, err
	}
	if err := util.CheckAddrValid(config.LocalAddr); err != nil {
		return nil, fmt.Errorf("invalid localAddr %v in config, error: %v", config.LocalAddr, err)
	}
	if err := s.AddProxyConfigToAgent(config.AgentId, remotePort, config.LocalAddr, config.IsWhiteListOn, config.WhiteListIps); err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}

func (h *rpcHandlers) ListProxyConfigs(ctx context.Context, input *pb.Empty) (*pb.ListProxyConfigsOutput, error) {
	s := anywhereServer.GetServerInstance()
	if s == nil {
		return nil, fmt.Errorf("anywhere server not init")
	}
	res := &pb.ListProxyConfigsOutput{
		Config: make([]*pb.ProxyConfig, 0),
	}
	configs := s.ListProxyConfigs()
	for _, config := range configs {
		res.Config = append(res.Config, &pb.ProxyConfig{
			AgentId:    config.AgentId,
			RemotePort: config.RemoteAddr,
			LocalAddr:  config.LocalAddr,
		})
	}
	return res, nil
}

func (h *rpcHandlers) RemoveProxyConfig(ctx context.Context, input *pb.RemoveProxyConfigInput) (*pb.Empty, error) {
	s := anywhereServer.GetServerInstance()
	if s == nil {
		return nil, fmt.Errorf("anywhere server not init")
	}
	return &pb.Empty{}, s.RemoveProxyConfigFromAgent(input.AgentId, input.LocalAddr)
}

func NewClient() (pb.AnywhereServerClient, error) {
	cc, err := grpc.Dial(grpcAddress.String(), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return pb.NewAnywhereServerClient(cc), nil
}

func newClientWithPort(port int) (pb.AnywhereServerClient, error) {
	addr, err := util.GetAddrByIpPort("127.0.0.1", port)
	if err != nil {
		return nil, err
	}
	cc, err := grpc.Dial(addr.String(), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return pb.NewAnywhereServerClient(cc), nil
}

func ListAgent(port int) error {
	client, err := newClientWithPort(port)
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

func AddProxyConfig(port int, agentId, remotePort, localAddr string, isWhiteListOn bool, whiteListIps string) error {

	client, err := newClientWithPort(port)
	if err != nil {
		return err
	}
	input := &pb.AddProxyConfigInput{Config: &pb.ProxyConfig{
		AgentId:       agentId,
		RemotePort:    remotePort,
		LocalAddr:     localAddr,
		IsWhiteListOn: isWhiteListOn,
		WhiteListIps:  whiteListIps,
	}}
	_, err = client.AddProxyConfig(context.Background(), input)
	if err != nil {
		return fmt.Errorf("add proxy config error: %v", err)
	}
	return nil

}

func ListProxyConfigs(port int) error {

	boolToString := func(b bool) string {
		if b {
			return "ON"
		}
		return "OFF"
	}
	client, err := newClientWithPort(port)
	if err != nil {
		return err
	}
	configs, err := client.ListProxyConfigs(context.Background(), &pb.Empty{})
	if err != nil {
		return fmt.Errorf("list proxy config error: %v", err)
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"AgentId", "RemoteAddr", "LocalAddr", "IsWhiteListOn", "IpWhiteList"})
	for _, config := range configs.Config {
		table.Append([]string{config.AgentId, "0.0.0.0:" + config.RemotePort, config.LocalAddr, boolToString(config.IsWhiteListOn), config.WhiteListIps})
	}
	table.Render()
	return nil

}

func RemoveProxyConfig(port int, agentId, localAddr string) error {

	client, err := newClientWithPort(port)
	if err != nil {
		return err
	}
	_, err = client.RemoveProxyConfig(context.Background(), &pb.RemoveProxyConfigInput{
		AgentId:   agentId,
		LocalAddr: localAddr,
	})
	return err

}

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
	pb.RegisterAnywhereServerServer(grpcServer, GetRpcHandlers())
	grpcAddress = addr
	if err := grpcServer.Serve(l); err != nil {
		errChan <- err
	}
}
