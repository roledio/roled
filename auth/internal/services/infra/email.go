package infra

import (
	"context"

	strip "github.com/grokify/html-strip-tags-go"
	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/models"
	"gopkg.in/gomail.v2"
)

type EmailService interface {
	Send(ctx context.Context, req models.SendEmailRequest) error
}

type emailService struct {
	defaultConfig *configs.DefaultConfig
	dialer        *gomail.Dialer
}

func NewEmailService(defaultConfig *configs.DefaultConfig) EmailService {
	gomail.SetPartEncoding(gomail.QuotedPrintable)
	dialer := gomail.NewDialer(
		defaultConfig.Email.SMTP.Host,
		int(defaultConfig.Email.SMTP.Port),
		defaultConfig.Email.SMTP.Username,
		defaultConfig.Email.SMTP.Password)
	return &emailService{
		defaultConfig: defaultConfig,
		dialer:        dialer,
	}
}

func (s *emailService) Send(ctx context.Context, req models.SendEmailRequest) error {
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
