package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/queues/handlers"
	"github.com/roledio/roled/auth/internal/queues/payloads"
	"github.com/roledio/roled/auth/pkg/email"
	em "github.com/roledio/roled/auth/pkg/email/mocks"
	rm "github.com/roledio/roled/auth/pkg/redis/mocks"
	"github.com/shomali11/util/xhashes"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEmailJobs(t *testing.T) {
	for _, kind := range []string{"reset_password", "activate_member", "verify_email", "invite_user"} {
		for _, outcome := range []string{"production", "development", "already suffixed", "invalid duration", "send error"} {
			t.Run(kind+"/"+outcome, func(t *testing.T) {
				ctx := context.Background()
				sender, redis := em.NewMockService(t), rm.NewMockService(t)
				config := &configs.DefaultConfig{Env: "development", BaseURL: "https://auth.example", ResetWithContextExpiryDuration: "1h", ActivateMemberExpiryDuration: "1h", VerifyEmailExpiryDuration: "1h"}
				if outcome == "production" {
					config.Env = "production"
				}
				if outcome == "invalid duration" {
					config.ResetWithContextExpiryDuration, config.ActivateMemberExpiryDuration, config.VerifyEmailExpiryDuration = "invalid", "invalid", "invalid"
				}
				login, logo := "https://app.example/login", "https://app.example/logo.png"
				payload := payloads.EmailPayload{Type: kind, From: "no-reply@example.com", To: "member@example.com", Subject: "Account action", AccountName: "Acme", ProjectName: "<script>Portal</script>", ProjectLogoURL: &logo, DisplayName: "<script>Member</script>", UserID: "user", LoginURL: &login, Token: "supplied-token", IsSignup: true}
				if outcome == "already suffixed" {
					payload.Subject += " #development"
				}
				var tokenKey string
				if kind == "verify_email" && outcome != "invalid duration" {
					redis.EXPECT().SetData(ctx, mock.Anything, models.EmailVerifyTokenData{UserID: "user", LoginURL: &login}, time.Hour).Run(func(_ context.Context, key string, _ any, _ time.Duration) { tokenKey = key }).Return(nil).Once()
				}
				failure := errors.New("SMTP unavailable")
				if outcome != "invalid duration" {
					var sendErr error
					if outcome == "send error" {
						sendErr = failure
					}
					sender.EXPECT().Send(ctx, mock.Anything).Run(func(_ context.Context, req email.Request) {
						require.Equal(t, payload.From, req.From)
						require.Equal(t, []string{payload.To}, req.To)
						require.True(t, req.IsHTML)
						require.Contains(t, req.Body, "Portal")
						require.Contains(t, req.Body, "&lt;script&gt;Portal&lt;/script&gt;")
						require.NotContains(t, req.Body, "<script>")
						if outcome == "production" {
							require.Equal(t, "Account action", req.Subject)
						} else {
							require.Equal(t, "Account action #development", req.Subject)
						}
						paths := map[string]string{"reset_password": "/password/reset/", "activate_member": "/member/activate/", "invite_user": "/user/activate/"}
						if kind == "verify_email" {
							matches := regexp.MustCompile(`https://auth\.example/email/verify/([A-Za-z0-9_-]+)`).FindStringSubmatch(req.Body)
							require.Len(t, matches, 2)
							require.Len(t, matches[1], 64)
							require.Equal(t, "email_verify:"+xhashes.SHA256(matches[1]), tokenKey)
						} else {
							require.Contains(t, req.Body, config.BaseURL+paths[kind]+payload.Token)
						}
					}).Return(sendErr).Once()
				}
				data, err := json.Marshal(payload)
				require.NoError(t, err)
				err = handlers.NewEmailHandler(config, sender, redis).Handle(ctx, string(data))
				switch outcome {
				case "invalid duration":
					require.Error(t, err)
				case "send error":
					require.ErrorIs(t, err, failure)
				default:
					require.NoError(t, err)
				}
			})
		}
	}
}

func TestEmailJobRejectsInvalidPayloadsAndTokenStorageFailure(t *testing.T) {
	for _, raw := range []string{"{", `{"type":"unknown"}`, `{"type":"verify_email","user_id":"user"}`} {
		t.Run(raw, func(t *testing.T) {
			sender, redis := em.NewMockService(t), rm.NewMockService(t)
			failure := errors.New("Redis unavailable")
			if strings.Contains(raw, "verify_email") {
				redis.EXPECT().SetData(mock.Anything, mock.Anything, mock.Anything, time.Hour).Return(failure).Once()
			}
			err := handlers.NewEmailHandler(&configs.DefaultConfig{VerifyEmailExpiryDuration: "1h"}, sender, redis).Handle(context.Background(), raw)
			require.Error(t, err)
			if strings.Contains(raw, "verify_email") {
				require.ErrorIs(t, err, failure)
			}
			sender.AssertNotCalled(t, "Send", mock.Anything, mock.Anything)
		})
	}
}
