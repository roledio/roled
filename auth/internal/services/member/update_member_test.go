package member

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.openly.dev/pointy"

	"github.com/roledio/roled/internal/constants"
	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/internal/errors"
	repositorymocks "github.com/roledio/roled/internal/mocks/repositories"
	"github.com/roledio/roled/internal/models"
	pkgerrors "github.com/roledio/roled/pkg/errors"
)

func TestMemberService_UpdateMember_SystemProjectNotFound(t *testing.T) {
	ctx := context.Background()

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(nil, nil)

	service := &memberService{registry: mockRegistry}

	req := &models.UpdateMemberRequest{
		MemberID: "member-id",
		IsAdmin:  pointy.Bool(true),
	}

	dbErr := fmt.Errorf("system project not found")
	res, err := service.UpdateMember(ctx, req)
	assert.Nil(t, res)
	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrSystemError.WithError(dbErr), err)
}

func TestMemberService_UpdateMember_AccessTokenNotFound(t *testing.T) {
	ctx := context.Background()
	systemProject := &entities.Project{ID: "sys-proj", IsActive: true}

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)

	service := &memberService{registry: mockRegistry}

	req := &models.UpdateMemberRequest{
		MemberID: "member-id",
		IsAdmin:  pointy.Bool(true),
	}

	res, err := service.UpdateMember(ctx, req)
	assert.Nil(t, res)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrCtxAccessTokenNotFound, err)
}

func TestMemberService_UpdateMember_AccountNotFound(t *testing.T) {
	ctx := context.Background()
	systemProject := &entities.Project{ID: "sys-proj", IsActive: true}
	accessToken := &entities.AccessToken{ProjectID: "sys-proj"}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)

	service := &memberService{registry: mockRegistry}

	req := &models.UpdateMemberRequest{
		MemberID: "member-id",
		IsAdmin:  pointy.Bool(true),
	}

	res, err := service.UpdateMember(ctx, req)
	assert.Nil(t, res)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrCtxAccountNotFound, err)
}

func TestMemberService_UpdateMember_MemberNotFound(t *testing.T) {
	ctx := context.Background()
	systemProject := &entities.Project{ID: "sys-proj", IsActive: true}
	accessToken := &entities.AccessToken{ProjectID: "sys-proj"}
	account := &entities.Account{ID: "acc-1", IsSystem: true}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)
	mockMemberRepo.EXPECT().FindByID(ctx, "member-id").Return(nil, nil)

	service := &memberService{registry: mockRegistry}

	req := &models.UpdateMemberRequest{
		MemberID: "member-id",
		IsAdmin:  pointy.Bool(true),
	}

	res, err := service.UpdateMember(ctx, req)
	assert.Nil(t, res)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrMemberNotFound, err)
}

func TestMemberService_UpdateMember_ClientJWT_NonSystemAccount_DifferentAccount(t *testing.T) {
	ctx := context.Background()
	systemProject := &entities.Project{ID: "sys-proj", IsActive: true}
	accessToken := &entities.AccessToken{ProjectID: "sys-proj", UserID: nil}
	account := &entities.Account{ID: "acc-1", IsSystem: false}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	targetMember := &entities.Member{ID: "member-id", AccountID: "acc-2", IsAdmin: false}

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)
	mockMemberRepo.EXPECT().FindByID(ctx, "member-id").Return(targetMember, nil)

	service := &memberService{registry: mockRegistry}

	req := &models.UpdateMemberRequest{
		MemberID: "member-id",
		IsAdmin:  pointy.Bool(true),
	}

	res, err := service.UpdateMember(ctx, req)
	assert.Nil(t, res)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrMemberNotFound, err)
}

func TestMemberService_UpdateMember_SystemAccount_AccountIDMismatch(t *testing.T) {
	ctx := context.Background()
	systemProject := &entities.Project{ID: "sys-proj", IsActive: true}
	accessToken := &entities.AccessToken{ProjectID: "sys-proj", UserID: nil}
	account := &entities.Account{ID: "acc-1", IsSystem: true}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	targetMember := &entities.Member{ID: "member-id", AccountID: "acc-2", IsAdmin: false}

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)
	mockMemberRepo.EXPECT().FindByID(ctx, "member-id").Return(targetMember, nil)

	service := &memberService{registry: mockRegistry}

	req := &models.UpdateMemberRequest{
		MemberID:  "member-id",
		AccountID: "acc-3",
		IsAdmin:   pointy.Bool(true),
	}

	res, err := service.UpdateMember(ctx, req)
	assert.Nil(t, res)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrMemberNotFound, err)
}

func TestMemberService_UpdateMember_ClientJWT_DemoteLastAdmin(t *testing.T) {
	ctx := context.Background()
	systemProject := &entities.Project{ID: "sys-proj", IsActive: true}
	accessToken := &entities.AccessToken{ProjectID: "sys-proj", UserID: nil}
	account := &entities.Account{ID: "acc-1", IsSystem: true}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	targetMember := &entities.Member{ID: "member-id", AccountID: "acc-1", IsAdmin: true}

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)
	mockMemberRepo.EXPECT().FindByID(ctx, "member-id").Return(targetMember, nil)
	mockMemberRepo.EXPECT().CountByAccountID(ctx, "acc-1", pointy.Bool(true)).Return(1, nil)

	service := &memberService{registry: mockRegistry}

	req := &models.UpdateMemberRequest{
		MemberID: "member-id",
		IsAdmin:  pointy.Bool(false),
	}

	res, err := service.UpdateMember(ctx, req)
	assert.Nil(t, res)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrCannotDemoteLastAdmin, err)
}

func TestMemberService_UpdateMember_ClientJWT_NoOp(t *testing.T) {
	ctx := context.Background()
	systemProject := &entities.Project{ID: "sys-proj", IsActive: true}
	accessToken := &entities.AccessToken{ProjectID: "sys-proj", UserID: nil}
	account := &entities.Account{ID: "acc-1", IsSystem: true}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	targetMember := &entities.Member{ID: "member-id", AccountID: "acc-1", IsAdmin: true, UserID: "user-id", CreatedAt: time.Now(), UpdatedAt: time.Now()}

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)
	mockMemberRepo.EXPECT().FindByID(ctx, "member-id").Return(targetMember, nil)

	service := &memberService{registry: mockRegistry}

	req := &models.UpdateMemberRequest{
		MemberID: "member-id",
		IsAdmin:  nil,
	}

	res, err := service.UpdateMember(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, targetMember.ID, res.ID)
	assert.Equal(t, targetMember.IsAdmin, res.IsAdmin)
}

func TestMemberService_UpdateMember_ClientJWT_UpdateSuccess(t *testing.T) {
	ctx := context.Background()
	systemProject := &entities.Project{ID: "sys-proj", IsActive: true}
	accessToken := &entities.AccessToken{ProjectID: "sys-proj", UserID: nil}
	account := &entities.Account{ID: "acc-1", IsSystem: true}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	targetMember := &entities.Member{ID: "member-id", AccountID: "acc-1", IsAdmin: false, UserID: "user-id", CreatedAt: time.Now(), UpdatedAt: time.Now()}

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)
	mockMemberRepo.EXPECT().FindByID(ctx, "member-id").Return(targetMember, nil)
	mockMemberRepo.EXPECT().Update(ctx, targetMember).Return(1, nil)

	service := &memberService{registry: mockRegistry}

	req := &models.UpdateMemberRequest{
		MemberID: "member-id",
		IsAdmin:  pointy.Bool(true),
	}

	res, err := service.UpdateMember(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, targetMember.ID, res.ID)
	assert.True(t, res.IsAdmin)
}

func TestMemberService_UpdateMember_UserJWT_UpdateSelf(t *testing.T) {
	ctx := context.Background()
	systemProject := &entities.Project{ID: "sys-proj", IsActive: true}
	accessToken := &entities.AccessToken{ProjectID: "sys-proj", UserID: pointy.String("current-user")}
	account := &entities.Account{ID: "acc-1", IsSystem: true}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	targetMember := &entities.Member{ID: "member-id", AccountID: "acc-1", IsAdmin: true, UserID: "current-user"}

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)
	mockMemberRepo.EXPECT().FindByID(ctx, "member-id").Return(targetMember, nil)

	service := &memberService{registry: mockRegistry}

	req := &models.UpdateMemberRequest{
		MemberID: "member-id",
		IsAdmin:  pointy.Bool(false),
	}

	res, err := service.UpdateMember(ctx, req)
	assert.Nil(t, res)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrCannotUpdateSelf, err)
}

func TestMemberService_UpdateMember_UserJWT_NonSystemAccount_CallerNotAdmin(t *testing.T) {
	ctx := context.Background()
	systemProject := &entities.Project{ID: "sys-proj", IsActive: true}
	accessToken := &entities.AccessToken{ProjectID: "sys-proj", UserID: pointy.String("current-user"), AccountID: "acc-1"}
	account := &entities.Account{ID: "acc-1", IsSystem: false}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	targetMember := &entities.Member{ID: "target-member-id", AccountID: "acc-1", IsAdmin: false, UserID: "other-user"}
	callerMember := &entities.Member{ID: "caller-member-id", AccountID: "acc-1", IsAdmin: false, UserID: "current-user"}

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)
	mockMemberRepo.EXPECT().FindByID(ctx, "target-member-id").Return(targetMember, nil)
	mockMemberRepo.EXPECT().FindByAccountIDAndUserID(ctx, "acc-1", "current-user").Return(callerMember, nil)

	service := &memberService{registry: mockRegistry}

	req := &models.UpdateMemberRequest{
		MemberID: "target-member-id",
		IsAdmin:  pointy.Bool(true),
	}

	res, err := service.UpdateMember(ctx, req)
	assert.Nil(t, res)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrNonAdminUpdateMember, err)
}

func TestMemberService_UpdateMember_UserJWT_NonSystemAccount_Success(t *testing.T) {
	ctx := context.Background()
	systemProject := &entities.Project{ID: "sys-proj", IsActive: true}
	accessToken := &entities.AccessToken{ProjectID: "sys-proj", UserID: pointy.String("current-user"), AccountID: "acc-1"}
	account := &entities.Account{ID: "acc-1", IsSystem: false}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	targetMember := &entities.Member{ID: "target-member-id", AccountID: "acc-1", IsAdmin: false, UserID: "other-user", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	callerMember := &entities.Member{ID: "caller-member-id", AccountID: "acc-1", IsAdmin: true, UserID: "current-user"}

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)
	mockMemberRepo.EXPECT().FindByID(ctx, "target-member-id").Return(targetMember, nil)
	mockMemberRepo.EXPECT().FindByAccountIDAndUserID(ctx, "acc-1", "current-user").Return(callerMember, nil)
	mockMemberRepo.EXPECT().Update(ctx, targetMember).Return(1, nil)

	service := &memberService{registry: mockRegistry}

	req := &models.UpdateMemberRequest{
		MemberID: "target-member-id",
		IsAdmin:  pointy.Bool(true),
	}

	res, err := service.UpdateMember(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "target-member-id", res.ID)
	assert.True(t, res.IsAdmin)
}

func TestMemberService_UpdateMember_UpdateByID_NoRowsAffected(t *testing.T) {
	ctx := context.Background()
	systemProject := &entities.Project{ID: "sys-proj", IsActive: true}
	accessToken := &entities.AccessToken{ProjectID: "sys-proj", UserID: nil}
	account := &entities.Account{ID: "acc-1", IsSystem: true}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	targetMember := &entities.Member{ID: "member-id", AccountID: "acc-1", IsAdmin: false, UserID: "user-id"}

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)
	mockMemberRepo.EXPECT().FindByID(ctx, "member-id").Return(targetMember, nil)
	mockMemberRepo.EXPECT().Update(ctx, targetMember).Return(0, nil)

	service := &memberService{registry: mockRegistry}

	req := &models.UpdateMemberRequest{
		MemberID: "member-id",
		IsAdmin:  pointy.Bool(true),
	}

	res, err := service.UpdateMember(ctx, req)
	assert.Nil(t, res)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrMemberNotFound, err)
}

func TestMemberService_UpdateMember_UpdateByID_Error(t *testing.T) {
	ctx := context.Background()
	systemProject := &entities.Project{ID: "sys-proj", IsActive: true}
	accessToken := &entities.AccessToken{ProjectID: "sys-proj", UserID: nil}
	account := &entities.Account{ID: "acc-1", IsSystem: true}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	targetMember := &entities.Member{ID: "member-id", AccountID: "acc-1", IsAdmin: false, UserID: "user-id"}
	dbErr := assert.AnError

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)
	mockMemberRepo.EXPECT().FindByID(ctx, "member-id").Return(targetMember, nil)
	mockMemberRepo.EXPECT().Update(ctx, targetMember).Return(0, dbErr)

	service := &memberService{registry: mockRegistry}

	req := &models.UpdateMemberRequest{
		MemberID: "member-id",
		IsAdmin:  pointy.Bool(true),
	}

	res, err := service.UpdateMember(ctx, req)
	assert.Nil(t, res)
	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrSystemError.WithError(dbErr), err)
}
