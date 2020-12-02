package persist

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWhiteList(t *testing.T) {
	Init("anywhere:anywhere@tcp(10.0.0.2:3306)/anywhere_test?charset=utf8mb4&parseTime=True&loc=Local")
	_, err := DB.Exec("truncate table whitelist_deny_history")
	assert.Equal(t, nil, err)
	assert.Equal(t, nil, AddWhiteListDenyIp(9495, "agent-1", "10.0.0.2:22", "8.8.8.8"))
	res, err := GetTotalDenyRank()
	assert.Equal(t, 1, len(res))
	assert.Equal(t, "8.8.8.8", res[0].Ip)
	assert.Equal(t, int64(1), res[0].Count)
}
