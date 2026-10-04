package api

import (
	"encoding/json"
	"github.com/gofiber/fiber/v3"
	"github.com/roledio/roled/auth/internal/models"
	sm "github.com/roledio/roled/auth/internal/services/project/mocks"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetProjectsHTTPContract(t *testing.T) {
	for _, outcome := range []string{"success", "service error"} {
		t.Run(outcome, func(t *testing.T) {
			service := sm.NewMockProjectService(t)
			var expected models.GetProjectsRequest
			expected.PageNum = 1
			expected.PageSize = 5
			require.NoError(t, json.Unmarshal([]byte(`{}`), &expected))
			result := []models.GetProjectsResponse{{ID: "p", Name: "Project"}}
			if outcome != "invalid binding" {
				var failure error
				if outcome == "service error" {
					failure = pkgerrors.ErrSystemError
				}
				call := service.On("GetProjects", mock.Anything, &expected).Once()
				call.Return(result, 7, failure)
			}
			app := fiber.New()
			app.Add([]string{"GET"}, "/projects", (&handler{projectService: service}).getProjects)
			body := ``
			url := "/projects?page_num=1&page_size=5"
			if outcome == "invalid binding" {
				url = "/projects?page_num=bad"
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
func TestGetProjectDetailsHTTPContract(t *testing.T) {
	for _, outcome := range []string{"success", "service error", "invalid binding"} {
		t.Run(outcome, func(t *testing.T) {
			service := sm.NewMockProjectService(t)
			var expected models.GetProjectDetailsRequest
			require.NoError(t, json.Unmarshal([]byte(`{}`), &expected))
			expected.ProjectID = "p"
			result := &models.ProjectDetails{ID: "p", Name: "Project"}
			if outcome != "invalid binding" {
				var failure error
				if outcome == "service error" {
					failure = pkgerrors.ErrSystemError
				}
				call := service.On("GetProjectDetails", mock.Anything, &expected).Once()
				call.Return(result, failure)
			}
			app := fiber.New()
			app.Add([]string{"GET"}, "/projects/:project_id", (&handler{projectService: service}).getProjectDetails)
			body := ``
			url := "/projects/p"
			if outcome == "invalid binding" {
				app = fiber.New()
				app.Add([]string{"GET"}, "/missing", (&handler{projectService: service}).getProjectDetails)
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
func TestCreateProjectHTTPContract(t *testing.T) {
	for _, outcome := range []string{"success", "service error", "invalid binding"} {
		t.Run(outcome, func(t *testing.T) {
			service := sm.NewMockProjectService(t)
			var expected models.CreateProjectRequest
			require.NoError(t, json.Unmarshal([]byte(`{"name":"Project"}`), &expected))
			result := &models.ProjectDetails{ID: "p", Name: "Project"}
			if outcome != "invalid binding" {
				var failure error
				if outcome == "service error" {
					failure = pkgerrors.ErrSystemError
				}
				call := service.On("CreateProject", mock.Anything, &expected).Once()
				call.Return(result, failure)
			}
			app := fiber.New()
			app.Add([]string{"POST"}, "/projects", (&handler{projectService: service}).createProject)
			body := `{"name":"Project"}`
			url := "/projects"
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
func TestUpdateProjectHTTPContract(t *testing.T) {
	for _, outcome := range []string{"success", "service error", "invalid binding"} {
		t.Run(outcome, func(t *testing.T) {
			service := sm.NewMockProjectService(t)
			var expected models.UpdateProjectRequest
			require.NoError(t, json.Unmarshal([]byte(`{"name":"Project","is_active":true}`), &expected))
			expected.ProjectID = "p"
			result := &models.ProjectDetails{ID: "p", Name: "Project"}
			if outcome != "invalid binding" {
				var failure error
				if outcome == "service error" {
					failure = pkgerrors.ErrSystemError
				}
				call := service.On("UpdateProject", mock.Anything, &expected).Once()
				call.Return(result, failure)
			}
			app := fiber.New()
			app.Add([]string{"PUT"}, "/projects/:project_id", (&handler{projectService: service}).updateProject)
			body := `{"name":"Project","is_active":true}`
			url := "/projects/p"
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
func TestDeleteProjectHTTPContract(t *testing.T) {
	for _, outcome := range []string{"success", "service error", "invalid binding"} {
		t.Run(outcome, func(t *testing.T) {
			service := sm.NewMockProjectService(t)
			var expected models.DeleteProjectRequest
			require.NoError(t, json.Unmarshal([]byte(`{"name":"Project"}`), &expected))
			expected.ProjectID = "p"
			if outcome != "invalid binding" {
				var failure error
				if outcome == "service error" {
					failure = pkgerrors.ErrSystemError
				}
				call := service.On("DeleteProject", mock.Anything, &expected).Once()
				call.Return(failure)
			}
			app := fiber.New()
			app.Add([]string{"DELETE"}, "/projects/:project_id", (&handler{projectService: service}).deleteProject)
			body := `{"name":"Project"}`
			url := "/projects/p"
			if outcome == "invalid binding" {
				app = fiber.New()
				app.Add([]string{"DELETE"}, "/missing", (&handler{projectService: service}).deleteProject)
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
func TestGetProjectSettingsHTTPContract(t *testing.T) {
	for _, outcome := range []string{"success", "service error", "invalid binding"} {
		t.Run(outcome, func(t *testing.T) {
			service := sm.NewMockProjectService(t)
			var expected models.GetProjectSettingsRequest
			require.NoError(t, json.Unmarshal([]byte(`{}`), &expected))
			expected.ProjectID = "p"
			result := &models.ProjectSettings{IsSignupEnabled: true}
			if outcome != "invalid binding" {
				var failure error
				if outcome == "service error" {
					failure = pkgerrors.ErrSystemError
				}
				call := service.On("GetProjectSettings", mock.Anything, &expected).Once()
				call.Return(result, failure)
			}
			app := fiber.New()
			app.Add([]string{"GET"}, "/projects/:project_id/settings", (&handler{projectService: service}).getProjectSettings)
			body := ``
			url := "/projects/p/settings"
			if outcome == "invalid binding" {
				app = fiber.New()
				app.Add([]string{"GET"}, "/missing", (&handler{projectService: service}).getProjectSettings)
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
func TestUpdateProjectSettingsHTTPContract(t *testing.T) {
	for _, outcome := range []string{"success", "service error", "invalid binding"} {
		t.Run(outcome, func(t *testing.T) {
			service := sm.NewMockProjectService(t)
			var expected models.UpdateProjectSettingsRequest
			require.NoError(t, json.Unmarshal([]byte(`{"is_signup_enabled":true,"is_signup_verify_email":false,"is_forgot_password_enabled":true,"is_allow_temp_email":false}`), &expected))
			expected.ProjectID = "p"
			result := &models.ProjectSettings{IsSignupEnabled: true}
			if outcome != "invalid binding" {
				var failure error
				if outcome == "service error" {
					failure = pkgerrors.ErrSystemError
				}
				call := service.On("UpdateProjectSettings", mock.Anything, &expected).Once()
				call.Return(result, failure)
			}
			app := fiber.New()
			app.Add([]string{"PUT"}, "/projects/:project_id/settings", (&handler{projectService: service}).updateProjectSettings)
			body := `{"is_signup_enabled":true,"is_signup_verify_email":false,"is_forgot_password_enabled":true,"is_allow_temp_email":false}`
			url := "/projects/p/settings"
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
func TestUpdateProjectSignupRoleHTTPContract(t *testing.T) {
	for _, outcome := range []string{"success", "service error", "invalid binding"} {
		t.Run(outcome, func(t *testing.T) {
			service := sm.NewMockProjectService(t)
			var expected models.UpdateProjectSignupRoleRequest
			require.NoError(t, json.Unmarshal([]byte(`{"role_id":"r"}`), &expected))
			expected.ProjectID = "p"
			result := &models.UpdateProjectSignupRoleResponse{RoleID: "r", RoleName: "Reader"}
			if outcome != "invalid binding" {
				var failure error
				if outcome == "service error" {
					failure = pkgerrors.ErrSystemError
				}
				call := service.On("UpdateProjectSignupRole", mock.Anything, &expected).Once()
				call.Return(result, failure)
			}
			app := fiber.New()
			app.Add([]string{"PUT"}, "/projects/:project_id/signup-role", (&handler{projectService: service}).updateProjectSignupRole)
			body := `{"role_id":"r"}`
			url := "/projects/p/signup-role"
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
