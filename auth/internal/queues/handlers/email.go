package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gofiber/fiber/v3/log"
	"github.com/karrick/tparse/v2"
	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/constants/rediskeys"
	"github.com/roledio/roled/auth/internal/mail"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/queues"
	"github.com/roledio/roled/auth/internal/queues/payloads"
	"github.com/roledio/roled/auth/internal/services/infra"
	"github.com/roledio/roled/auth/pkg/utils/idutil"
	"github.com/shomali11/util/xhashes"
)

type emailHandler struct {
	defaultConfig *configs.DefaultConfig
	emailService  infra.EmailService
	redisService  infra.RedisService
}

func NewEmailHandler(defaultConfig *configs.DefaultConfig, emailService infra.EmailService, redisService infra.RedisService) queues.Handler {
	return emailHandler{
		defaultConfig: defaultConfig,
		emailService:  emailService,
		redisService:  redisService,
	}
}

func (h emailHandler) Handle(ctx context.Context, payload string) error {
	log.WithContext(ctx).Debugw("Processing email job", "payload", payload)
	var p payloads.EmailPayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		return err
	}
	switch p.Type {
	case constants.EmailPayloadTypeResetPassword:
		return h.handleResetPassword(ctx, p)
	case constants.EmailPayloadTypeActivateMember:
		return h.handleActivateMember(ctx, p)
	case constants.EmailPayloadTypeVerifyEmail:
		return h.handleVerificationEmail(ctx, p)
	case constants.EmailPayloadTypeInviteUser:
		return h.handleInviteUser(ctx, p)
	default:
		return fmt.Errorf("unsupported email type: %s", p.Type)
	}
}

func (h emailHandler) handleResetPassword(ctx context.Context, payload payloads.EmailPayload) error {
	now := time.Now()
	duration, err := tparse.AbsoluteDuration(now, h.defaultConfig.ResetWithContextExpiryDuration)
	if err != nil {
		return err
	}
	expiresIn := humanize.RelTime(now, now.Add(duration), "", "")
	resetPasswordURL := fmt.Sprintf("%s/password/reset/%s", h.defaultConfig.BaseURL, payload.Token)
	body, err := mail.LoadAndParseTemplate("templates/html/reset-password.html", map[string]any{
		"ProjectName":      payload.ProjectName,
		"ProjectLogoURL":   payload.ProjectLogoURL,
		"ProjectIsSystem":  payload.ProjectIsSystem,
		"DisplayName":      payload.DisplayName,
		"Year":             now.Year(),
		"ExpiresIn":        strings.TrimSpace(expiresIn),
		"ResetPasswordURL": resetPasswordURL,
	})
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to load email template", "error", err)
		return err
	}
	return h.send(ctx, payload.From, payload.To, payload.Subject, body)
}

func (h emailHandler) handleActivateMember(ctx context.Context, payload payloads.EmailPayload) error {
	now := time.Now()
	duration, err := tparse.AbsoluteDuration(now, h.defaultConfig.ActivateMemberExpiryDuration)
	if err != nil {
		return err
	}
	expiresIn := humanize.RelTime(now, now.Add(duration), "", "")
	activateMemberURL := fmt.Sprintf("%s/member/activate/%s", h.defaultConfig.BaseURL, payload.Token)
	body, err := mail.LoadAndParseTemplate("templates/html/activate-member.html", map[string]any{
		"AccountName":       payload.AccountName,
		"ProjectName":       payload.ProjectName,
		"ProjectLogoURL":    payload.ProjectLogoURL,
		"ProjectIsSystem":   payload.ProjectIsSystem,
		"DisplayName":       payload.DisplayName,
		"Year":              now.Year(),
		"ExpiresIn":         strings.TrimSpace(expiresIn),
		"ActivateMemberURL": activateMemberURL,
	})
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to load email template", "error", err)
		return err
	}
	return h.send(ctx, payload.From, payload.To, payload.Subject, body)
}

func (h emailHandler) handleVerificationEmail(ctx context.Context, payload payloads.EmailPayload) error {

	tokenData := models.EmailVerifyTokenData{
		UserID:   payload.UserID,
		LoginURL: payload.LoginURL,
	}
	now := time.Now()
	token := idutil.NanoID(64)
	tokenHash := xhashes.SHA256(token)
	tokenExpiryDuration, err := tparse.AbsoluteDuration(now, h.defaultConfig.VerifyEmailExpiryDuration)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to parse verify email expiry duration", "duration", h.defaultConfig.VerifyEmailExpiryDuration, "error", err)
		return err
	}
	// Store token hash in Redis with expiration
	redisKey := fmt.Sprintf("%s:%s", rediskeys.EmailVerifyPrefix, tokenHash)
	err = h.redisService.SetData(ctx, redisKey, tokenData, tokenExpiryDuration)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to store email verification token in redis", "error", err)
		return err
	}

	expiresIn := humanize.RelTime(now, now.Add(tokenExpiryDuration), "", "")
	verifyURL := fmt.Sprintf("%s/email/verify/%s", h.defaultConfig.BaseURL, token)
	body, err := mail.LoadAndParseTemplate("templates/html/verify-email.html", map[string]any{
		"ProjectName":     payload.ProjectName,
		"ProjectLogoURL":  payload.ProjectLogoURL,
		"ProjectIsSystem": payload.ProjectIsSystem,
		"VerifyURL":       verifyURL,
		"Year":            now.Year(),
		"ExpiresIn":       strings.TrimSpace(expiresIn),
		"LoginURL":        payload.LoginURL,
		"DisplayName":     payload.DisplayName,
		"IsSignup":        payload.IsSignup,
	})
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to load email template", "error", err)
		return err
	}
	return h.send(ctx, payload.From, payload.To, payload.Subject, body)
}

func (h emailHandler) handleInviteUser(ctx context.Context, payload payloads.EmailPayload) error {
	now := time.Now()
	duration, err := tparse.AbsoluteDuration(now, h.defaultConfig.ActivateMemberExpiryDuration)
	if err != nil {
		return err
	}
	expiresIn := humanize.RelTime(now, now.Add(duration), "", "")
	activateProjectUserURL := fmt.Sprintf("%s/user/activate/%s", h.defaultConfig.BaseURL, payload.Token)
	body, err := mail.LoadAndParseTemplate("templates/html/invite-user.html", map[string]any{
		"AccountName":            payload.AccountName,
		"ProjectName":            payload.ProjectName,
		"ProjectLogoURL":         payload.ProjectLogoURL,
		"ProjectIsSystem":        payload.ProjectIsSystem,
		"DisplayName":            payload.DisplayName,
		"Year":                   now.Year(),
		"ExpiresIn":              strings.TrimSpace(expiresIn),
		"ActivateProjectUserURL": activateProjectUserURL,
	})
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to load email template", "error", err)
		return err
	}
	return h.send(ctx, payload.From, payload.To, payload.Subject, body)
}

func (h emailHandler) send(ctx context.Context, from, to, subject, body string) error {
	req := models.SendEmailRequest{
		From:    from,
		To:      []string{to},
		Subject: h.appendEnv(subject),
		Body:    body,
		IsHTML:  true,
	}
	err := h.emailService.Send(ctx, req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to send email", "error", err, "from", from, "to", to, "subject", subject)
		return err
	}
	log.WithContext(ctx).Debugw("Successfully sent email", "from", from, "to", to, "subject", subject)
	return nil
}

func (h emailHandler) appendEnv(subject string) string {
	env := fmt.Sprintf("#%s", h.defaultConfig.Env)
	if !h.defaultConfig.IsEnvProd() && !strings.HasSuffix(subject, env) {
		return fmt.Sprintf("%s %s", subject, env)
	}
	return subject
}
