package tool

import (
	"crypto/tls"

	"github.com/cntechpower/anywhere/server/conf"

	"gopkg.in/gomail.v2"
)

func Send(subject string, toAddress []string) error {
	c := conf.Conf.SmtpConfig
	d := gomail.NewDialer(c.Host, c.Port, c.UserName, c.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	m := gomail.NewMessage()
	m.SetHeader("From", c.UserName)
	m.SetHeader("To", toAddress...)
	m.SetAddressHeader("Cc", c.UserName, "Anywhere")
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", "Hello <b>Bob</b> and <i>Cora</i>!")
	return d.DialAndSend(m)
	// Send emails using d.
}
