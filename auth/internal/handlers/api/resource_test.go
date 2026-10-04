package api

import (
	"encoding/json"
	"github.com/gofiber/fiber/v3"
	"github.com/roledio/roled/auth/internal/models"
	sm "github.com/roledio/roled/auth/internal/services/resource/mocks"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetResourcesHTTPContract(t *testing.T) {
	for _, outcome := range []string{"success", "service error", "invalid binding"} {
		t.Run(outcome, func(t *testing.T) {
			service := sm.NewMockResourceService(t)
			var expected models.GetResourcesRequest
			expected.PageNum = 1
			expected.PageSize = 5
			require.NoError(t, json.Unmarshal([]byte(`{}`), &expected))
			expected.ProjectID = "p"
			result := []models.ResourceDetails{{ID: "r", Name: "Documents"}}
			if outcome != "invalid binding" {
				var failure error
				if outcome == "service error" {
					failure = pkgerrors.ErrSystemError
				}
				call := service.On("GetResources", mock.Anything, &expected).Once()
				call.Return(result, 7, failure)
			}
			app := fiber.New()
			app.Add([]string{"GET"}, "/projects/:project_id/resources", (&handler{resourceService: service}).getProjectResources)
			body := ``
			url := "/projects/p/resources?page_num=1&page_size=5"
			if outcome == "invalid binding" {
				app = fiber.New()
				app.Add([]string{"GET"}, "/missing", (&handler{resourceService: service}).getProjectResources)
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
func TestGetResourceDetailsHTTPContract(t *testing.T) {
	for _, outcome := range []string{"success", "service error", "invalid binding"} {
		t.Run(outcome, func(t *testing.T) {
			service := sm.NewMockResourceService(t)
			var expected models.GetResourceDetailsRequest
			require.NoError(t, json.Unmarshal([]byte(`{}`), &expected))
			expected.ProjectID = "p"
			expected.ResourceID = "r"
			result := &models.ResourceDetails{ID: "r", Name: "Documents"}
			if outcome != "invalid binding" {
				var failure error
				if outcome == "service error" {
					failure = pkgerrors.ErrSystemError
				}
				call := service.On("GetResourceDetails", mock.Anything, &expected).Once()
				call.Return(result, failure)
			}
			app := fiber.New()
			app.Add([]string{"GET"}, "/projects/:project_id/resources/:resource_id", (&handler{resourceService: service}).getResourceDetails)
			body := ``
			url := "/projects/p/resources/r"
			if outcome == "invalid binding" {
				app = fiber.New()
				app.Add([]string{"GET"}, "/missing", (&handler{resourceService: service}).getResourceDetails)
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
func TestCreateResourceHTTPContract(t *testing.T) {
	for _, outcome := range []string{"success", "service error", "invalid binding"} {
		t.Run(outcome, func(t *testing.T) {
			service := sm.NewMockResourceService(t)
			var expected models.CreateResourceRequest
			require.NoError(t, json.Unmarshal([]byte(`{"name":"Documents","code":"documents"}`), &expected))
			expected.ProjectID = "p"
			result := &models.ResourceDetails{ID: "r", Name: "Documents"}
			if outcome != "invalid binding" {
				var failure error
				if outcome == "service error" {
					failure = pkgerrors.ErrSystemError
				}
				call := service.On("CreateResource", mock.Anything, &expected).Once()
				call.Return(result, failure)
			}
			app := fiber.New()
			app.Add([]string{"POST"}, "/projects/:project_id/resources", (&handler{resourceService: service}).createProjectResource)
			body := `{"name":"Documents","code":"documents"}`
			url := "/projects/p/resources"
			if outcome == "invalid binding" {
				body = "{"
			}
			req := httptest.NewRequest("POST", url, strings.NewReader(body))
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
func TestUpdateResourceHTTPContract(t *testing.T) {
	for _, outcome := range []string{"success", "service error", "invalid binding"} {
		t.Run(outcome, func(t *testing.T) {
			service := sm.NewMockResourceService(t)
			var expected models.UpdateResourceRequest
			require.NoError(t, json.Unmarshal([]byte(`{"name":"Documents","code":"documents"}`), &expected))
			expected.ProjectID = "p"
			expected.ResourceID = "r"
			result := &models.ResourceDetails{ID: "r", Name: "Documents"}
			if outcome != "invalid binding" {
				var failure error
				if outcome == "service error" {
					failure = pkgerrors.ErrSystemError
				}
				call := service.On("UpdateResource", mock.Anything, &expected).Once()
				call.Return(result, failure)
			}
			app := fiber.New()
			app.Add([]string{"PUT"}, "/projects/:project_id/resources/:resource_id", (&handler{resourceService: service}).updateProjectResource)
			body := `{"name":"Documents","code":"documents"}`
			url := "/projects/p/resources/r"
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
func TestDeleteResourceHTTPContract(t *testing.T) {
	for _, outcome := range []string{"success", "service error", "invalid binding"} {
		t.Run(outcome, func(t *testing.T) {
			service := sm.NewMockResourceService(t)
			var expected models.DeleteResourceRequest
			require.NoError(t, json.Unmarshal([]byte(`{}`), &expected))
			expected.ProjectID = "p"
			expected.ResourceID = "r"
			if outcome != "invalid binding" {
				var failure error
				if outcome == "service error" {
					failure = pkgerrors.ErrSystemError
				}
				call := service.On("DeleteResource", mock.Anything, &expected).Once()
				call.Return(failure)
			}
			app := fiber.New()
			app.Add([]string{"DELETE"}, "/projects/:project_id/resources/:resource_id", (&handler{resourceService: service}).deleteProjectResource)
			body := ``
			url := "/projects/p/resources/r"
			if outcome == "invalid binding" {
				app = fiber.New()
				app.Add([]string{"DELETE"}, "/missing", (&handler{resourceService: service}).deleteProjectResource)
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
