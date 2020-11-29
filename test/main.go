package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cntechpower/anywhere/test/client"
	"github.com/cntechpower/anywhere/test/client/operations"
)

var addr string
var o operations.ClientService
var agentId string
var userName string
var errBuilder strings.Builder

func init() {
	flag.StringVar(&addr, "addr", "127.0.0.1:1114", "rest api address")
	o = client.NewHTTPClientWithConfig(nil, &client.TransportConfig{
		Host:     addr,
		BasePath: client.DefaultBasePath,
		Schemes:  client.DefaultSchemes,
	}).Operations
}

func getDefaultContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	return ctx, cancel
}

func addError(funcName string, err error) {
	errBuilder.WriteString(fmt.Sprintf("calling %s error: %v\n", funcName, err.Error()))
}

func addProxyConfig(remotePort int64, localAddr string) error {
	ctx, cancel := getDefaultContext()
	_, err := o.PostV1ProxyAdd(&operations.PostV1ProxyAddParams{
		UserName:        userName,
		AgentID:         agentId,
		LocalAddr:       localAddr,
		RemotePort:      remotePort,
		WhiteListEnable: false,
		WhiteListIps:    nil,
		Context:         ctx,
	})
	cancel()
	if err != nil {
		addError("PostV1ProxyAdd", err)
		return err
	}
	return nil
}

func main() {
	flag.Parse()

	// check agent online
	{
		ctx, cancel := getDefaultContext()
		resp, err := o.GetV1AgentList(operations.NewGetV1AgentListParamsWithContext(ctx))
		cancel()
		if err != nil {
			addError("GetV1AgentList", err)
			goto END
		}
		if len(resp.Payload) != 1 {
			addError("GetV1AgentList", fmt.Errorf("length of agent list should be 1"))
			goto END
		}
		for _, agent := range resp.Payload {
			agentId = agent.AgentID
			userName = agent.UserName
		}
	}
	// add proxy config and check
	{
		//MySQL 8.0.19
		if err := addProxyConfig(4444, "172.90.101.21:3306"); err != nil {
			goto END
		}
		//MySQL 5.7.28
		if err := addProxyConfig(4445, "172.90.101.22:3306"); err != nil {
			goto END
		}
		//Nginx
		if err := addProxyConfig(4446, "172.90.101.23:80"); err != nil {
			goto END
		}
		//SSH
		if err := addProxyConfig(4447, "172.90.101.24:22"); err != nil {
			goto END
		}
		ctx, cancel := getDefaultContext()
		resp, err := o.GetV1ProxyList(operations.NewGetV1ProxyListParamsWithContext(ctx))
		cancel()
		if err != nil {
			addError("GetV1ProxyList", err)
			goto END
		}
		if len(resp.Payload) != 4 {
			addError("GetV1ProxyList", fmt.Errorf("length of proxy list should be 4"))
		}

	}

END:
	errMsg := errBuilder.String()
	if errMsg != "" {
		fmt.Println(errMsg)
		os.Exit(1)
	}
	fmt.Println("test success")
	os.Exit(0)
}
