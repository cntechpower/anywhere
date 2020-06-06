package anywhereServer

import (
	"anywhere/model"
	"math/rand"
	"testing"
	"time"
)

func printProxyConfig(t *testing.T, list []*model.ProxyConfig) {
	t.Log("-------------------------------------")
	for _, p := range list {
		t.Log(p.NetworkFlowRemoteToLocalInBytes)
	}
	t.Log("-------------------------------------")
}

func checkSort(t *testing.T, list []*model.ProxyConfig) {
	for i := 1; i < len(list); i++ {
		if list[i].NetworkFlowRemoteToLocalInBytes > list[i-1].NetworkFlowRemoteToLocalInBytes {
			t.Errorf("list[%v]-- %v > list[%v] -- %v", i, list[i], i-1, list[i-1])
		}
	}
}
func TestServer_Cache(t *testing.T) {
	origin := make([]*model.ProxyConfig, 0)
	for i := 15; i > 0; i-- {
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
