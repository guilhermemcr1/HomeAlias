package alert

import (
	"fmt"
	"net/smtp"
)

type SMTP struct {
	Host string
	Port int
	User string
	Pass string
	From string
}

func (s *SMTP) Send(to, subject, body string) error {
	if s.Host == "" {
		return fmt.Errorf("smtp not configured")
	}
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	msg := []byte("To: " + to + "\r\nSubject: " + subject + "\r\n\r\n" + body + "\r\n")
	var auth smtp.Auth
	if s.User != "" {
		auth = smtp.PlainAuth("", s.User, s.Pass, s.Host)
	}
	return smtp.SendMail(addr, auth, s.From, []string{to}, msg)
}
