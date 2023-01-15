package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWhiteList(t *testing.T) {
	l, err := NewWhiteList(1, "1", "1", "", true)
	if !assert.Equal(t, nil, err) {
		t.FailNow()
	}

	for i := 0; i <= connCountRejectCount; i++ {
		assert.Equal(t, true, l.IpInWhiteList(nil, "127.0.0.1"))
	}
	assert.Equal(t, false, l.IpInWhiteList(nil, "127.0.0.1"))

	time.Sleep(connCountRefreshInterval*2)

	assert.Equal(t, true, l.IpInWhiteList(nil, "127.0.0.1"))
}
