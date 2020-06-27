package model

import (
	"fmt"
	"strings"
)

type ProxyConfigHeap struct {
	maxLength int
	list      []*ProxyConfig
	less      func(i, j *ProxyConfig) bool
}

func (ph *ProxyConfigHeap) Push(config *ProxyConfig) bool {
	n := len(ph.list)
	if n > ph.maxLength {
		panic("heap length overflow")
	}
	if n == ph.maxLength && ph.less(config, ph.list[0]) {
		return false
	}
	if n < ph.maxLength {
		ph.list = append(ph.list, config)
		ph.up(len(ph.list) - 1)
	} else { //len(ph.list) == ph.maxLength && !ph.less(config, ph.list[0])
		ph.list[0] = config
		ph.down(0, len(ph.list))
	}
	return true
}

func (ph *ProxyConfigHeap) IsEmpty() bool {
	return len(ph.list) == 0
}

func (ph *ProxyConfigHeap) IsValid() bool {
	n := len(ph.list)
	if n == 1 {
		return true
	}
	for i := n/2 - 1; i >= 0; i-- {
		if ph.less(ph.list[2*i+1], ph.list[i]) || (2*i+2 < n && ph.less(ph.list[2*i+2], ph.list[i])) {
			fmt.Printf("invalid list[%v]: %v,list[%v]: %v,list[%v]: %v,\n", i, ph.list[i], 2*i+1, ph.list[2*i+1], 2*i+2, ph.list[2*i+2])
			return false
		}
	}
	return true
}
func (ph *ProxyConfigHeap) Pop() *ProxyConfig {
	n := len(ph.list) - 1
	ph.list[0], ph.list[n] = ph.list[n], ph.list[0]
	ph.down(0, len(ph.list)-1)
	c := ph.list[n]
	ph.list = ph.list[0:n]
	//fmt.Printf("checking valid of ph: %v\n", ph.IsValid())
	return c

}
func (ph *ProxyConfigHeap) Len() int {
	return len(ph.list)
}

func (ph *ProxyConfigHeap) down(i int, stop int) bool {
	iCopy := i
	for {
		leftNodeIdx := 2*iCopy + 1
		if leftNodeIdx >= stop || leftNodeIdx < 0 {
			//no below nodes or overflow
			break
		}
		smallerNodeIdx := leftNodeIdx
		rightNodeIdx := leftNodeIdx + 1
		if rightNodeIdx < stop && ph.less(ph.list[rightNodeIdx], ph.list[leftNodeIdx]) {
			smallerNodeIdx = rightNodeIdx
		}
		if !ph.less(ph.list[smallerNodeIdx], ph.list[iCopy]) {
			//below smaller node it not smaller than parent
			break
		}
		ph.list[iCopy], ph.list[smallerNodeIdx] = ph.list[smallerNodeIdx], ph.list[iCopy]
		iCopy = smallerNodeIdx
	}
	return iCopy > i
}

func (ph *ProxyConfigHeap) debugPrint() string {
	s := strings.Builder{}
	s.WriteString("[")
	for _, c := range ph.list {
		s.WriteString(fmt.Sprintf("%v,", c.NetworkFlowRemoteToLocalInBytes))
	}
	s.WriteString("]")
	return s.String()
}

func (ph *ProxyConfigHeap) up(i int) bool {
	iCopy := i
	for {
		parentIdx := (iCopy - 1) / 2
		if parentIdx == iCopy || !ph.less(ph.list[iCopy], ph.list[parentIdx]) {
			return iCopy < i
		}
		ph.list[iCopy], ph.list[parentIdx] = ph.list[parentIdx], ph.list[iCopy]
		iCopy = parentIdx
	}
}
func InitProxyConfigHeap(cs []*ProxyConfig, less func(i, j *ProxyConfig) bool, limit int) *ProxyConfigHeap {
	ph := &ProxyConfigHeap{
		list:      make([]*ProxyConfig, 0, limit+1),
		less:      less,
		maxLength: limit,
	}
	for _, c := range cs {
		ph.Push(c)
		//fmt.Printf("after push %v to heap result %v, heap is %v\n", c.NetworkFlowRemoteToLocalInBytes, ok, ph.debugPrint())
	}
	//fmt.Printf("checking valid of ph: %v\n", ph.IsValid())
	return ph
}
