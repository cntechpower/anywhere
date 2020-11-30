package tool

import (
	"testing"

	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/server/conf"
	"github.com/stretchr/testify/assert"
)

func TestSendMail(t *testing.T) {
	conf.Conf = &model.SystemConfig{}
	conf.Conf.SmtpConfig = &model.SmtpConfig{
		Host:     "smtp.exmail.qq.com",
		Port:     465,
		UserName: "no_reply@cntechpower.com",
		Password: "APB0K77gamkkAaFc",
	}
	assert.Equal(t, nil, Send("test", []string{"root@cntechpower.com"}))
}
