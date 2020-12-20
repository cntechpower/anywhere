package persist

import (
	"testing"

	"github.com/cntechpower/anywhere/log"

	"github.com/stretchr/testify/assert"
)

func TestWhiteList(t *testing.T) {
	log.InitLogger("")
	Init("anywhere:anywhere@tcp(10.0.0.2:3306)/anywhere_test?charset=utf8mb4&parseTime=True&loc=Local&readTimeout=5s&timeout=5s")
	_, err := DB.Exec("truncate table whitelist_deny_history")
	assert.Equal(t, nil, err)
	assert.Equal(t, nil, AddWhiteListDenyIp(9495, "agent-1", "10.0.0.2:22", "8.8.8.8"))
	assert.Equal(t, nil, AddWhiteListDenyIp(9495, "agent-1", "10.0.0.2:22", "8.8.8.9"))
	assert.Equal(t, nil, AddWhiteListDenyIp(9495, "agent-1", "10.0.0.2:23", "8.8.8.9"))
	assert.Equal(t, nil, err)
	res, c, ic, err := GetWhiteListDenyRank("total", 1)
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(res))
	assert.Equal(t, int64(3), c)
	assert.Equal(t, int64(2), ic)
	assert.Equal(t, "8.8.8.9", res[0].Ip)
	assert.Equal(t, int64(2), res[0].Count)
	_, err = DB.Exec("drop table whitelist_deny_history")
	assert.Equal(t, nil, err)
}
