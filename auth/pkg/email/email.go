package email

import (
	"context"

	strip "github.com/grokify/html-strip-tags-go"
	"gopkg.in/gomail.v2"
)

type Request struct {
	To      []string
	CC      []string
	BCC     []string
	From    string
	Subject string
	Body    string
	IsHTML  bool
	Sender  string // Optional Sender header to set in the email
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

type Service interface {
	Send(ctx context.Context, req Request) error
}

type service struct {
	dialer *gomail.Dialer
}

func NewService(config *SMTPConfig) Service {
	gomail.SetPartEncoding(gomail.QuotedPrintable)
	dialer := gomail.NewDialer(
		config.Host,
		config.Port,
		config.Username,
		config.Password)
	return &service{
		dialer: dialer,
	}
}

func (s *service) Send(ctx context.Context, req Request) error {
	m := gomail.NewMessage()

	if req.Sender != "" {
		// Set the Sender header to avoid (GitLab) SMTP error 550: email rejected and classified as SPAM
		// https://stackoverflow.com/questions/11055481/send-smtp-with-from-address-of-another-domain/11055765#11055765
		m.SetHeader("Sender", req.Sender)
	}

	m.SetHeader("From", req.From)
	m.SetHeader("To", req.To...)
	m.SetHeader("Subject", req.Subject)

	if len(req.CC) > 0 {
		m.SetHeader("Cc", req.CC...)
	}
	if len(req.BCC) > 0 {
		m.SetHeader("Bcc", req.BCC...)
	}

	if req.IsHTML {
		m.SetBody("text/plain", strip.StripTags(req.Body))
		m.AddAlternative("text/html", req.Body)
	} else {
		m.SetBody("text/plain", strip.StripTags(req.Body))
	}

	return s.dialer.DialAndSend(m)
}
