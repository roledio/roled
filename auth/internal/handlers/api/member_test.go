package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/roledio/roled/internal/configs"
	customerrors "github.com/roledio/roled/internal/errors"
	servicemocks "github.com/roledio/roled/internal/mocks/services"
	"github.com/roledio/roled/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tidwall/gjson"
	"go.openly.dev/pointy"
)

func TestUpdateMember_Success(t *testing.T) {
	app := fiber.New()
	memberServiceMock := servicemocks.NewMockMemberService(t)

	now := time.Now()
	expectedResponse := &models.UpdateMemberResponse{
		ID:        "member-id",
		AccountID: "account-id",
		UserID:    "user-id",
		IsAdmin:   true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	memberServiceMock.EXPECT().UpdateMember(mock.Anything, &models.UpdateMemberRequest{
		MemberID: "member-id",
		IsAdmin:  pointy.Bool(true),
	}).Return(expectedResponse, nil)

	deps := &Dependencies{
		MemberService: memberServiceMock,
	}
	h := NewHandler(app, &configs.DefaultConfig{}, deps)

	app.Patch("/api/v1/members/:member_id", h.updateMember)

	requestBody, _ := json.Marshal(map[string]interface{}{
		"is_admin": true,
	})

	req := httptest.NewRequest("PATCH", "/api/v1/members/member-id", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	res, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)

	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	parsed := gjson.Parse(body)

	assert.True(t, parsed.Get("success").Bool())
	assert.Equal(t, "member-id", parsed.Get("data.id").String())
	assert.Equal(t, "account-id", parsed.Get("data.account_id").String())
	assert.Equal(t, "user-id", parsed.Get("data.user_id").String())
	assert.True(t, parsed.Get("data.is_admin").Bool())
}

func TestUpdateMember_ValidationError(t *testing.T) {
	app := fiber.New()
	memberServiceMock := servicemocks.NewMockMemberService(t)

	deps := &Dependencies{
		MemberService: memberServiceMock,
	}
	h := NewHandler(app, &configs.DefaultConfig{}, deps)

	app.Patch("/api/v1/members/:member_id", h.updateMember)

	// Sending account_id that is too short to trigger struct validation error
	requestBody, _ := json.Marshal(map[string]interface{}{
		"account_id": "short",
	})

	req := httptest.NewRequest("PATCH", "/api/v1/members/member-id", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	res, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, res.StatusCode)

	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	parsed := gjson.Parse(body)

	assert.False(t, parsed.Get("success").Bool())
	assert.NotEmpty(t, parsed.Get("error.message").String())
}

func TestUpdateMember_ServiceError(t *testing.T) {
	app := fiber.New()
	memberServiceMock := servicemocks.NewMockMemberService(t)

	memberServiceMock.EXPECT().UpdateMember(mock.Anything, &models.UpdateMemberRequest{
		MemberID: "member-id",
		IsAdmin:  pointy.Bool(false),
	}).Return(nil, customerrors.ErrCannotDemoteLastAdmin)

	deps := &Dependencies{
		MemberService: memberServiceMock,
	}
	h := NewHandler(app, &configs.DefaultConfig{}, deps)

	app.Patch("/api/v1/members/:member_id", h.updateMember)

	requestBody, _ := json.Marshal(map[string]interface{}{
		"is_admin": false,
	})

	req := httptest.NewRequest("PATCH", "/api/v1/members/member-id", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	res, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, res.StatusCode)

	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	parsed := gjson.Parse(body)

	assert.False(t, parsed.Get("success").Bool())
	assert.Equal(t, customerrors.ErrCannotDemoteLastAdmin.Code, parsed.Get("error.code").String())
	assert.Equal(t, customerrors.ErrCannotDemoteLastAdmin.Msg, parsed.Get("error.message").String())
}
