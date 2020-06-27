package anywhereServer

import (
	"anywhere/model"
	"math/rand"
	"testing"
	"time"
)

const proxyConfigLength = 10000
const silence = true

func printProxyConfig(t *testing.T, list []*model.ProxyConfig) {
	if silence {
		return
	}
	t.Log("-------------------------------------")
	for _, p := range list {
		t.Log(p.NetworkFlowRemoteToLocalInBytes)
	}
	t.Log("-------------------------------------")
}

type testingLogger interface {
	Errorf(format string, args ...interface{})
}

func checkSort(l testingLogger, list []*model.ProxyConfig) {
	for i := 1; i < len(list); i++ {
		if list[i].NetworkFlowRemoteToLocalInBytes > list[i-1].NetworkFlowRemoteToLocalInBytes {
			l.Errorf("list[%v]-- %v > list[%v] -- %v", i, list[i], i-1, list[i-1])
		}
	}
}
func TestServer_Cache(t *testing.T) {
	origin := make([]*model.ProxyConfig, 0)
	for i := proxyConfigLength; i > 0; i-- {
		origin = append(origin, &model.ProxyConfig{
			NetworkFlowRemoteToLocalInBytes: rand.New(rand.NewSource(time.Now().UnixNano())).Uint64(),
		})
	}
	t.Logf("Before sorting, len is %v", len(origin))
	printProxyConfig(t, origin)
	after := SortDescAndLimit(origin, func(p1 *model.ProxyConfig, p2 *model.ProxyConfig) bool {
		return p1.NetworkFlowRemoteToLocalInBytes < p2.NetworkFlowRemoteToLocalInBytes
	}, 10)
	t.Logf("After sorting, len is %v", len(after))
	printProxyConfig(t, after)
	checkSort(t, after)
}

func TestServer_Cache_Heap(t *testing.T) {
	origin := make([]*model.ProxyConfig, 0)
	for i := proxyConfigLength; i > 0; i-- {
		origin = append(origin, &model.ProxyConfig{
			NetworkFlowRemoteToLocalInBytes: rand.New(rand.NewSource(time.Now().UnixNano())).Uint64(),
		})
	}
	t.Logf("Before sorting, len is %v", len(origin))
	printProxyConfig(t, origin)
	after := SortDescAndLimitUsingHeap(origin, func(p1 *model.ProxyConfig, p2 *model.ProxyConfig) bool {
		return p1.NetworkFlowRemoteToLocalInBytes > p2.NetworkFlowRemoteToLocalInBytes
	}, 10)
	t.Logf("After sorting, len is %v", len(after))
	printProxyConfig(t, after)
	checkSort(t, after)
}

//GOPATH=/Users/dujinyang/code/cntechpower/anywhere/ go test -run=^$ -bench=BenchmarkServer_Cache  -benchtime="3s" -cpuprofile profile_cpu.out
func BenchmarkServer_Cache(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		origin := make([]*model.ProxyConfig, 0)
		for i := proxyConfigLength; i > 0; i-- {
			origin = append(origin, &model.ProxyConfig{
				NetworkFlowRemoteToLocalInBytes: rand.New(rand.NewSource(time.Now().UnixNano())).Uint64(),
			})
		}
		res := SortDescAndLimit(origin, func(p1 *model.ProxyConfig, p2 *model.ProxyConfig) bool {
			return p1.NetworkFlowRemoteToLocalInBytes < p2.NetworkFlowRemoteToLocalInBytes
		}, 10)
		checkSort(b, res)
	}
}

//go test -run=BenchmarkServer_Cache_Heap -bench=. -benchtime="3s" -cpuprofile profile_cpu.out
func BenchmarkServer_Heap_Cache(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		origin := make([]*model.ProxyConfig, 0)
		for i := proxyConfigLength; i > 0; i-- {
			origin = append(origin, &model.ProxyConfig{
				NetworkFlowRemoteToLocalInBytes: rand.New(rand.NewSource(time.Now().UnixNano())).Uint64(),
			})
		}
		res := SortDescAndLimitUsingHeap(origin, func(p1 *model.ProxyConfig, p2 *model.ProxyConfig) bool {
			return p1.NetworkFlowRemoteToLocalInBytes > p2.NetworkFlowRemoteToLocalInBytes
		}, 10)
		checkSort(b, res)
	}
}
