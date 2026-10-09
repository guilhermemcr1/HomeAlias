package alert

import (
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
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
	msg, err := buildEmail(s.From, to, subject, body)
	if err != nil {
		return err
	}
	// MIME headers may include display names; the SMTP envelope needs addresses only.
	sender, _ := mail.ParseAddress(s.From)
	recipient, _ := mail.ParseAddress(to)
	addr := net.JoinHostPort(s.Host, strconv.Itoa(s.Port))
	var auth smtp.Auth
	if s.User != "" {
		auth = smtp.PlainAuth("", s.User, s.Pass, s.Host)
	}
	return smtp.SendMail(addr, auth, sender.Address, []string{recipient.Address}, msg)
}
