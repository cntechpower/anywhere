package handler

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cntechpower/utils/log"
	"github.com/olekukonko/tablewriter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/cntechpower/anywhere/server/api/rpc/definitions"
	"github.com/cntechpower/anywhere/server/conf"
	"github.com/cntechpower/anywhere/server/server"
	tlsutil "github.com/cntechpower/anywhere/tls"
	"github.com/cntechpower/anywhere/util"
)

// ensureConf loads anywhered.json on demand for CLI commands.
func ensureConf() error {
	if conf.Conf != nil {
		return nil
	}
	return conf.Init()
}

// grpcAddr resolves the management gRPC address lazily. Reading the config in
// a package init() ran before conf.Init and failed silently.
func grpcAddr() string {
	if conf.Conf != nil && conf.Conf.UiConfig != nil {
		return conf.Conf.UiConfig.GrpcAddr
	}
	addr, _ := conf.GetGrpcAddr()
	return addr
}

func StartRpcServer(ctx context.Context, s *server.Server, addr string, tlsConfig *tls.Config, errChan chan error) {
	if err := util.CheckAddrValid(addr); err != nil {
		errChan <- err
		return
	}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		errChan <- err
		return
	}
	grpcServer := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)))
	pb.RegisterAnywhereServerServer(grpcServer, GetRpcHandlers(s))
	go func() {
		<-ctx.Done()
		stopped := make(chan struct{})
		go func() {
			grpcServer.GracefulStop()
			close(stopped)
		}()
		select {
		case <-stopped:
		case <-time.After(5 * time.Second):
			grpcServer.Stop()
		}
	}()
	if err := grpcServer.Serve(l); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		errChan <- err
	}
}

func NewClient(silenceOutput bool) (pb.AnywhereServerClient, error) {
	if err := ensureConf(); err != nil {
		return nil, err
	}
	h := log.NewHeader("grpc")
	opts := []grpc.DialOption{
		grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{},
			cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			if !silenceOutput {
				log.Infof(h, "calling %+v", method)
			}

			err := invoker(ctx, method, req, reply, cc, opts...)
			if !silenceOutput {
				log.Infof(h, "called %+v, error: %+v", method, err)
			}
			return err
		}),
	}

	if conf.Conf.UiConfig.GrpcInsecure {
		opts = append(opts, grpc.WithInsecure())
	} else {
		clientTLS, err := tlsutil.ParseClientTlsConfig(
			conf.Conf.AgentSsl.CertFile, conf.Conf.AgentSsl.KeyFile, conf.Conf.AgentSsl.CaFile,
			"anywhere.local", false)
		if err != nil {
			return nil, fmt.Errorf("build grpc client tls config: %w", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(clientTLS)))
	}

	cc, err := grpc.Dial(grpcAddr(), opts...)
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
	table := tablewriter.NewTable(os.Stdout)
	table.Header("UserName", "ZoneName", "AgentId", "AgentAddr", "LastAckSend", "LastAckRcv")
	for _, agent := range res.Agent {
		table.Append(agent.UserName, agent.ZoneName, agent.Id, agent.RemoteAddr, agent.LastAckSend, agent.LastAckRcv)
	}
	_ = table.Render()
	return nil
}

func AddProxyConfig(userName, zoneName string, remotePort int, localAddr string, isWhiteListOn bool, whiteListIps string) error {

	client, err := NewClient(true)
	if err != nil {
		return err
	}
	input := &pb.AddProxyConfigInput{Config: &pb.ProxyConfig{
		Username:      userName,
		ZoneName:      zoneName,
		RemotePort:    int64(remotePort),
		LocalAddr:     localAddr,
		IsWhiteListOn: isWhiteListOn,
		WhiteCidrList: whiteListIps,
	}}
	if _, err = client.AddProxyConfig(context.Background(), input); err != nil {
		return fmt.Errorf("add proxy config error: %+v", err)
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
		return fmt.Errorf("list proxy config error: %+v", err)
	}
	table := tablewriter.NewTable(os.Stdout)
	table.Header("UserName", "ZoneName", "ServerPort", "LocalAddr", "IsWhiteListOn", "IpWhiteList", "ConnectCount", "ConnectRejectCount", "TotalNetFlowsInMB")
	for _, config := range configs.Config {
		table.Append(
			config.Username, config.ZoneName, strconv.FormatInt(config.RemotePort, 10),
			config.LocalAddr, util.BoolToString(config.IsWhiteListOn), config.WhiteCidrList,
			strconv.FormatInt(config.ProxyConnectCount, 10), strconv.FormatInt(config.ProxyConnectRejectCount, 10),
			strconv.FormatFloat(float64(config.NetworkFlowRemoteToLocalInBytes+config.NetworkFlowLocalToRemoteInBytes)/1024/1024, 'f', 5, 64),
		)
	}
	_ = table.Render()
	return nil
}

func RemoveProxyConfig(configId int64) error {
	client, err := NewClient(true)
	if err != nil {
		return err
	}
	_, err = client.RemoveProxyConfig(context.Background(), &pb.RemoveProxyConfigInput{
		Id: configId,
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

func UpdateProxyConfigWhiteList(userName, zoneName, localAddr, whiteCidrs string, whiteListEnable bool) error {
	client, err := NewClient(true)
	if err != nil {
		return err
	}
	_, err = client.UpdateProxyConfigWhiteList(context.Background(), &pb.UpdateProxyConfigWhiteListInput{
		UserName:        userName,
		ZoneName:        zoneName,
		LocalAddr:       localAddr,
		WhiteCidrs:      whiteCidrs,
		WhiteListEnable: whiteListEnable,
	})
	return err
}

func ListConns(zoneName string) error {
	client, err := NewClient(true)
	if err != nil {
		return err
	}
	res, err := client.ListConnections(context.Background(), &pb.ListConnsInput{
		ZoneName: zoneName,
	})
	if err != nil {
		return err
	}
	if len(res.Conn) == 0 {
		fmt.Println("no conn exist")
		return nil
	}
	table := tablewriter.NewTable(os.Stdout)
	table.Header("UserName", "ZoneName", "AgentId", "ConnId", "SrcRemoteAddr", "SrcLocalAddr", "DstRemoteAddr", "DstLocalAddr")
	for _, conn := range res.Conn {
		table.Append(conn.UserName, conn.ZoneName, conn.AgentId, strconv.Itoa(int(conn.ConnId)), conn.SrcRemoteAddr, conn.SrcLocalAddr, conn.DstRemoteAddr, conn.DstLocalAddr)
	}
	_ = table.Render()
	return nil
}

func KillConn(id int64) error {
	client, err := NewClient(true)
	if err != nil {
		return err
	}
	_, err = client.KillConnById(context.Background(), &pb.KillConnByIdInput{
		Id: id,
	})
	return err
}

func FlushConnections() error {
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
	_, err = client.KillAllConnections(context.Background(), &pb.Empty{})
	return err
}
