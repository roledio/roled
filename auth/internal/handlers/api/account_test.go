package api

import (
	"encoding/json"
	"github.com/gofiber/fiber/v3"
	"github.com/roledio/roled/auth/internal/models"
	sm "github.com/roledio/roled/auth/internal/services/account/mocks"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetAccountsHTTPContract(t *testing.T) {
	for _, outcome := range []string{"success", "service error"} {
		t.Run(outcome, func(t *testing.T) {
			service := sm.NewMockAccountService(t)
			var expected models.GetAccountsRequest
			expected.PageNum = 1
			expected.PageSize = 5
			require.NoError(t, json.Unmarshal([]byte(`{}`), &expected))
			result := []models.GetAccountsResponse{{ID: "a", Name: "Account"}}
			if outcome != "invalid binding" {
				var failure error
				if outcome == "service error" {
					failure = pkgerrors.ErrSystemError
				}
				call := service.On("GetAccounts", mock.Anything, &expected).Once()
				call.Return(result, 7, failure)
			}
			app := fiber.New()
			app.Add([]string{"GET"}, "/accounts", (&handler{accountService: service}).getAccounts)
			body := ``
			url := "/accounts?page_num=1&page_size=5"
			if outcome == "invalid binding" {
				url = "/accounts?page_num=bad"
			}
			req := httptest.NewRequest("GET", url, strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			res, err := app.Test(req)
			require.NoError(t, err)
			defer func() { require.NoError(t, res.Body.Close()) }()
			var envelope struct {
				Success bool            `json:"success"`
				Data    json.RawMessage `json:"data"`
				Error   *struct {
					Code string `json:"code"`
				} `json:"error"`
				Pagination *struct {
					TotalData int `json:"total_data"`
				} `json:"pagination"`
			}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&envelope))
			switch outcome {
			case "success":
				require.Equal(t, 200, res.StatusCode)
				require.True(t, envelope.Success)
				want, err := json.Marshal(result)
				require.NoError(t, err)
				require.JSONEq(t, string(want), string(envelope.Data))
				require.NotNil(t, envelope.Pagination)
				require.Equal(t, 7, envelope.Pagination.TotalData)
			case "service error":
				require.Equal(t, 500, res.StatusCode)
				require.False(t, envelope.Success)
				require.NotNil(t, envelope.Error)
			case "invalid binding":
				require.Equal(t, 400, res.StatusCode)
				require.False(t, envelope.Success)
			}
		})
	}
}
func TestGetAccountDetailsHTTPContract(t *testing.T) {
	for _, outcome := range []string{"success", "service error", "invalid binding"} {
		t.Run(outcome, func(t *testing.T) {
			service := sm.NewMockAccountService(t)
			var expected models.GetAccountDetailsRequest
			require.NoError(t, json.Unmarshal([]byte(`{}`), &expected))
			expected.AccountID = "a"
			result := &models.GetAccountDetailsResponse{ID: "a"}
			if outcome != "invalid binding" {
				var failure error
				if outcome == "service error" {
					failure = pkgerrors.ErrSystemError
				}
				call := service.On("GetAccountDetails", mock.Anything, &expected).Once()
				call.Return(result, failure)
			}
			app := fiber.New()
			app.Add([]string{"GET"}, "/accounts/:account_id", (&handler{accountService: service}).getAccountDetails)
			body := ``
			url := "/accounts/a"
			if outcome == "invalid binding" {
				app = fiber.New()
				app.Add([]string{"GET"}, "/missing", (&handler{accountService: service}).getAccountDetails)
				url = "/missing"
			}
			req := httptest.NewRequest("GET", url, strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			res, err := app.Test(req)
			require.NoError(t, err)
			defer func() { require.NoError(t, res.Body.Close()) }()
			var envelope struct {
				Success bool            `json:"success"`
				Data    json.RawMessage `json:"data"`
				Error   *struct {
					Code string `json:"code"`
				} `json:"error"`
				Pagination *struct {
					TotalData int `json:"total_data"`
				} `json:"pagination"`
			}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&envelope))
			switch outcome {
			case "success":
				require.Equal(t, 200, res.StatusCode)
				require.True(t, envelope.Success)
				want, err := json.Marshal(result)
				require.NoError(t, err)
				require.JSONEq(t, string(want), string(envelope.Data))
			case "service error":
				require.Equal(t, 500, res.StatusCode)
				require.False(t, envelope.Success)
				require.NotNil(t, envelope.Error)
			case "invalid binding":
				require.Equal(t, 400, res.StatusCode)
				require.False(t, envelope.Success)
			}
		})
	}
}
func TestUpdateAccountHTTPContract(t *testing.T) {
	for _, outcome := range []string{"success", "service error", "invalid binding"} {
		t.Run(outcome, func(t *testing.T) {
			service := sm.NewMockAccountService(t)
			var expected models.UpdateAccountRequest
			require.NoError(t, json.Unmarshal([]byte(`{"name":"Account","is_active":true}`), &expected))
			expected.AccountID = "a"
			result := &models.UpdateAccountResponse{ID: "a", Name: "Account"}
			if outcome != "invalid binding" {
				var failure error
				if outcome == "service error" {
					failure = pkgerrors.ErrSystemError
				}
				call := service.On("UpdateAccount", mock.Anything, &expected).Once()
				call.Return(result, failure)
			}
			app := fiber.New()
			app.Add([]string{"PUT"}, "/accounts/:account_id", (&handler{accountService: service}).updateAccount)
			body := `{"name":"Account","is_active":true}`
			url := "/accounts/a"
			if outcome == "invalid binding" {
				body = "{"
			}
			req := httptest.NewRequest("PUT", url, strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			res, err := app.Test(req)
			require.NoError(t, err)
			defer func() { require.NoError(t, res.Body.Close()) }()
			var envelope struct {
				Success bool            `json:"success"`
				Data    json.RawMessage `json:"data"`
				Error   *struct {
					Code string `json:"code"`
				} `json:"error"`
				Pagination *struct {
					TotalData int `json:"total_data"`
				} `json:"pagination"`
			}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&envelope))
			switch outcome {
			case "success":
				require.Equal(t, 200, res.StatusCode)
				require.True(t, envelope.Success)
				want, err := json.Marshal(result)
				require.NoError(t, err)
				require.JSONEq(t, string(want), string(envelope.Data))
			case "service error":
				require.Equal(t, 500, res.StatusCode)
				require.False(t, envelope.Success)
				require.NotNil(t, envelope.Error)
			case "invalid binding":
				require.Equal(t, 400, res.StatusCode)
				require.False(t, envelope.Success)
			}
		})
	}
}
func TestDeleteAccountHTTPContract(t *testing.T) {
	for _, outcome := range []string{"success", "service error", "invalid binding"} {
		t.Run(outcome, func(t *testing.T) {
			service := sm.NewMockAccountService(t)
			var expected models.DeleteAccountRequest
			require.NoError(t, json.Unmarshal([]byte(`{}`), &expected))
			expected.AccountID = "a"
			if outcome != "invalid binding" {
				var failure error
				if outcome == "service error" {
					failure = pkgerrors.ErrSystemError
				}
				call := service.On("DeleteAccount", mock.Anything, &expected).Once()
				call.Return(failure)
			}
			app := fiber.New()
			app.Add([]string{"DELETE"}, "/accounts/:account_id", (&handler{accountService: service}).deleteAccount)
			body := `{}`
			url := "/accounts/a"
			if outcome == "invalid binding" {
				app = fiber.New()
				app.Add([]string{"DELETE"}, "/missing", (&handler{accountService: service}).deleteAccount)
				url = "/missing"
			}
			req := httptest.NewRequest("DELETE", url, strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			res, err := app.Test(req)
			require.NoError(t, err)
			defer func() { require.NoError(t, res.Body.Close()) }()
			var envelope struct {
				Success bool            `json:"success"`
				Data    json.RawMessage `json:"data"`
				Error   *struct {
					Code string `json:"code"`
				} `json:"error"`
				Pagination *struct {
					TotalData int `json:"total_data"`
				} `json:"pagination"`
			}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&envelope))
			switch outcome {
			case "success":
				require.Equal(t, 200, res.StatusCode)
				require.True(t, envelope.Success)
			case "service error":
				require.Equal(t, 500, res.StatusCode)
				require.False(t, envelope.Success)
				require.NotNil(t, envelope.Error)
			case "invalid binding":
				require.Equal(t, 400, res.StatusCode)
				require.False(t, envelope.Success)
			}
		})
	}
}
