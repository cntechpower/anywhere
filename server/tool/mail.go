package tool

import (
	"crypto/tls"

	"github.com/cntechpower/anywhere/server/conf"

	"gopkg.in/gomail.v2"
)

func Send(toAddress []string, subject, body string) error {
	c := conf.Conf.SmtpConfig
	d := gomail.NewDialer(c.Host, c.Port, c.UserName, c.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	m := gomail.NewMessage()
	m.SetHeader("From", c.UserName)
	m.SetHeader("To", toAddress...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	return d.DialAndSend(m)
}
