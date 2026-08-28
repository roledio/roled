package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/pkg/utils/requestutil"
	"github.com/roledio/roled/pkg/utils/responseutil"
)

func (h *handler) getMembers(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.GetMembersRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	members, total, err := h.memberService.GetMembers(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	pagination := responseutil.Paginate(req.PageRequest, len(members), total)
	return responseutil.SendSuccessWithPagination(c, members, pagination)
}

func (h *handler) getMemberDetails(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.GetMemberDetailsRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	member, err := h.memberService.GetMemberDetails(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, member)
}

func (h *handler) createMember(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.CreateMemberRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	member, err := h.memberService.CreateMember(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, member)
}

func (h *handler) updateMember(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.UpdateMemberRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	member, err := h.memberService.UpdateMember(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, member)
}

func (h *handler) deleteMember(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.DeleteMemberRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	err = h.memberService.DeleteMember(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, nil)
}
