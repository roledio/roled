package web

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/models"
	bm "github.com/roledio/roled/auth/internal/services/branding/mocks"
	sm "github.com/roledio/roled/auth/internal/services/member/mocks"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRenderActivateMemberTemplateAndFailures(t *testing.T) {
	for _, stage := range []string{"success", "invalid", "service error", "branding error"} {
		t.Run(stage, func(t *testing.T) {
			app := fiber.New(fiber.Config{Views: captureViews{}})
			service := sm.NewMockMemberService(t)
			branding := bm.NewMockService(t)
			h := &handler{defaultConfig: &configs.DefaultConfig{}, app: app, memberService: service, brandingService: branding}
			if stage != "invalid" {
				var err error
				if stage == "service error" {
					err = pkgerrors.ErrSystemError
				}
				service.On("RenderActivateMember", mock.Anything, &models.RenderActivateMemberRequest{Token: "token"}).Return(&models.RenderActivateMemberResponse{Project: &entities.Project{ID: "project", Name: "Project"}}, err).Once()
				if err == nil {
					var be error
					if stage == "branding error" {
						be = pkgerrors.ErrSystemError
					}
					branding.On("ResolveBranding", mock.Anything, mock.Anything).Return(&models.BrandingDetails{}, be).Once()
				}
			}
			route, url, body := "/activate/:token", "/activate/token", ""
			if stage == "invalid" {
				route = "/invalid"
				url = "/invalid"
				body = ""
			}
			app.Add([]string{"GET"}, route, h.renderActivateMember)
			request := httptest.NewRequest("GET", url, strings.NewReader(body))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			response, err := app.Test(request)
			require.NoError(t, err)
			defer func() { require.NoError(t, response.Body.Close()) }()
			if stage == "branding error" {
				require.Equal(t, 500, response.StatusCode)
				return
			}
			require.Equal(t, 200, response.StatusCode)
			var rendered struct {
				Name string
				Data json.RawMessage
			}
			require.NoError(t, json.NewDecoder(response.Body).Decode(&rendered))
			require.Contains(t, rendered.Name, "activate-member")
			require.NotEmpty(t, rendered.Data)
		})
	}
}
func TestSubmitActivateMemberTemplateAndFailures(t *testing.T) {
	for _, stage := range []string{"success", "invalid", "service error", "branding error"} {
		t.Run(stage, func(t *testing.T) {
			app := fiber.New(fiber.Config{Views: captureViews{}})
			service := sm.NewMockMemberService(t)
			branding := bm.NewMockService(t)
			h := &handler{defaultConfig: &configs.DefaultConfig{}, app: app, memberService: service, brandingService: branding}
			if stage != "invalid" {
				var err error
				if stage == "service error" {
					err = pkgerrors.ErrSystemError
				}
				service.On("SubmitActivateMember", mock.Anything, &models.SubmitActivateMemberRequest{Token: "token", DisplayName: "Example", Password: "passphrase", PasswordConfirmation: "passphrase"}).Return(&models.SubmitActivateMemberResponse{Project: &entities.Project{ID: "project", Name: "Project"}}, err).Once()
				if err == nil {
					var be error
					if stage == "branding error" {
						be = pkgerrors.ErrSystemError
					}
					branding.On("ResolveBranding", mock.Anything, mock.Anything).Return(&models.BrandingDetails{}, be).Once()
				}
			}
			route, url, body := "/activate/:token", "/activate/token", "display_name=Example&password=passphrase&password_confirmation=passphrase"
			if stage == "invalid" {
				route = "/invalid"
				url = "/invalid"
				body = ""
			}
			app.Add([]string{"POST"}, route, h.submitActivateMember)
			request := httptest.NewRequest("POST", url, strings.NewReader(body))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			response, err := app.Test(request)
			require.NoError(t, err)
			defer func() { require.NoError(t, response.Body.Close()) }()
			if stage == "branding error" {
				require.Equal(t, 500, response.StatusCode)
				return
			}
			if stage == "invalid" || stage == "service error" {
				require.Equal(t, 303, response.StatusCode)
				require.NotEmpty(t, response.Header.Values("Set-Cookie"))
				return
			}
			require.Equal(t, 200, response.StatusCode)
			var rendered struct {
				Name string
				Data json.RawMessage
			}
			require.NoError(t, json.NewDecoder(response.Body).Decode(&rendered))
			require.Contains(t, rendered.Name, "activate-member-success")
			require.NotEmpty(t, rendered.Data)
		})
	}
}
