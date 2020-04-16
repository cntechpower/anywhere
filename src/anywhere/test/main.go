package main

import (
	"anywhere/test/client"
	"anywhere/test/client/operations"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

var addr string

func init() {
	flag.StringVar(&addr, "addr", "127.0.0.1:1114", "rest api address")
}

func getDefaultContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	return ctx, cancel
}

func main() {
	flag.Parse()
	o := client.NewHTTPClientWithConfig(nil, &client.TransportConfig{
		Host:     addr,
		BasePath: client.DefaultBasePath,
		Schemes:  client.DefaultSchemes,
	}).Operations
	var errBuilder strings.Builder
	agentId := ""
	addError := func(funcName string, err error) {
		errBuilder.WriteString(fmt.Sprintf("calling %s error: %v\n", funcName, err.Error()))
	}
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
		}
	}
	// add proxy config and check
	{
		ctx, cancel := getDefaultContext()
		_, err := o.PostV1ProxyAdd(&operations.PostV1ProxyAddParams{
			AgentID:         agentId,
			LocalAddr:       "10.0.0.2:3306",
			RemotePort:      4444,
			WhiteListEnable: false,
			WhiteListIps:    nil,
			Context:         ctx,
		})
		cancel()
		if err != nil {
			addError("PostV1ProxyAdd", err)
			goto END
		}
		ctx, cancel = getDefaultContext()
		_, err = o.PostV1ProxyAdd(&operations.PostV1ProxyAddParams{
			AgentID:         agentId,
			LocalAddr:       "10.0.0.8:22",
			RemotePort:      4445,
			WhiteListEnable: false,
			WhiteListIps:    nil,
			Context:         ctx,
		})
		cancel()
		if err != nil {
			addError("PostV1ProxyAdd", err)
			goto END
		}
		ctx, cancel = getDefaultContext()
		resp, err := o.GetV1ProxyList(operations.NewGetV1ProxyListParamsWithContext(ctx))
		cancel()
		if err != nil {
			addError("GetV1ProxyList", err)
			goto END
		}
		if len(resp.Payload) != 2 {
			addError("GetV1ProxyList", fmt.Errorf("length of agent list should be 2"))
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
