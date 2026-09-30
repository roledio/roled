package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	domainerrors "github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	bm "github.com/roledio/roled/auth/internal/services/branding/mocks"
	pkgmodels "github.com/roledio/roled/auth/pkg/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBrandingHandlers(t *testing.T) {
	disabled := false
	for _, method := range []string{http.MethodGet, http.MethodPut} {
		for _, outcome := range []string{"success", "service failure", "invalid request"} {
			t.Run(method+"/"+outcome, func(t *testing.T) {
				service := bm.NewMockService(t)
				h := &handler{brandingService: service}
				app := fiber.New()
				var request any = &models.GetBrandingRequest{ProjectID: "project-id"}
				serviceMethod := "GetBranding"
				var body string
				handle := h.getProjectBranding
				if method == http.MethodPut {
					serviceMethod, handle = "UpdateBranding", h.updateProjectBranding
					body = `{"primary_color":"#112233","rounding":"sharp","enable_shadow":false,"enable_border":false}`
					request = &models.UpdateBrandingRequest{ProjectID: "project-id", PrimaryColor: "#112233", Rounding: "sharp", EnableShadow: &disabled, EnableBorder: &disabled}
				}
				route := "/projects/:project_id/branding"
				url := "/projects/project-id/branding"
				if outcome == "invalid request" {
					if method == http.MethodGet {
						route, url = "/branding", "/branding"
					} else {
						body = `{"primary_color":"invalid"}`
					}
				} else {
					var serviceErr error
					if outcome == "service failure" {
						serviceErr = domainerrors.ErrProjectNotFound
					}
					service.On(serviceMethod, mock.Anything, request).Return(&models.BrandingDetails{ProjectID: "project-id", PrimaryColor: "#112233", Rounding: "sharp"}, serviceErr).Once()
				}
				app.Add([]string{method}, route, handle)
				req := httptest.NewRequest(method, url, strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				res, err := app.Test(req)
				require.NoError(t, err)
				defer func() { require.NoError(t, res.Body.Close()) }()
				var response struct {
					Success bool                    `json:"success"`
					Data    *models.BrandingDetails `json:"data"`
					Error   *pkgmodels.ErrorBody    `json:"error"`
				}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&response))
				switch outcome {
				case "success":
					require.Equal(t, http.StatusOK, res.StatusCode)
					require.True(t, response.Success)
					require.Equal(t, &models.BrandingDetails{ProjectID: "project-id", PrimaryColor: "#112233", Rounding: "sharp"}, response.Data)
				case "service failure":
					require.Equal(t, http.StatusNotFound, res.StatusCode)
					require.False(t, response.Success)
					require.Nil(t, response.Data)
					require.Equal(t, domainerrors.ErrProjectNotFound.Code, response.Error.Code)
				case "invalid request":
					require.Equal(t, http.StatusBadRequest, res.StatusCode)
					require.False(t, response.Success)
					require.NotNil(t, response.Error)
					service.AssertNotCalled(t, serviceMethod, mock.Anything, mock.Anything)
				}
			})
		}
	}
}
