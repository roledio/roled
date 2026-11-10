package middlewares

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/models"
	repositorymocks "github.com/roledio/roled/auth/internal/repositories/mocks"
	"github.com/roledio/roled/auth/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestJWTRejectsTokenWithoutAudience(t *testing.T) {
	const signingKey = "test-signing-key"
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &models.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "token-id",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		ClientID: "client-id",
		UserID:   "user-id",
	}).SignedString([]byte(signingKey))
	assert.NoError(t, err)

	defaultConfig := &configs.DefaultConfig{}
	defaultConfig.JWT.SigningKey = signingKey
	app := fiber.New()
	app.Get("/protected", JWT(defaultConfig, repositorymocks.NewMockRegistry(t)), testHandler)

	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response, err := app.Test(request)

	assert.NoError(t, err)
	assert.Equal(t, errors.ErrInvalidAuthorizationToken.HttpCode, response.StatusCode)

	body, err := io.ReadAll(response.Body)
	assert.NoError(t, err)
	parsed := gjson.ParseBytes(body)
	assert.False(t, parsed.Get("success").Bool())
	assert.Equal(t, errors.ErrInvalidAuthorizationToken.Code, parsed.Get("error.code").String())
}
