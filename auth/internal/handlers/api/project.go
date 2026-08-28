package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/pkg/utils/requestutil"
	"github.com/roledio/roled/pkg/utils/responseutil"
)

func (h *handler) getProjects(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.GetProjectsRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	projects, total, err := h.projectService.GetProjects(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	pagination := responseutil.Paginate(req.PageRequest, len(projects), total)
	return responseutil.SendSuccessWithPagination(c, projects, pagination)
}

func (h *handler) getProjectDetails(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.GetProjectDetailsRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	project, err := h.projectService.GetProjectDetails(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, project)
}

func (h *handler) createProject(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.CreateProjectRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	project, err := h.projectService.CreateProject(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, project)
}

func (h *handler) updateProject(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.UpdateProjectRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	project, err := h.projectService.UpdateProject(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, project)
}

func (h *handler) deleteProject(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.DeleteProjectRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	err = h.projectService.DeleteProject(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, nil)
}

func (h *handler) getProjectSettings(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.GetProjectSettingsRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	settings, err := h.projectService.GetProjectSettings(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, settings)
}

func (h *handler) updateProjectSettings(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.UpdateProjectSettingsRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	settings, err := h.projectService.UpdateProjectSettings(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, settings)
}

func (h *handler) updateProjectSignupRole(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.UpdateProjectSignupRoleRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	settings, err := h.projectService.UpdateProjectSignupRole(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, settings)
}
