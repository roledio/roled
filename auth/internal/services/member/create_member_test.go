package member

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/constants/rediskeys"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/queues"
	qm "github.com/roledio/roled/auth/internal/queues/mocks"
	"github.com/roledio/roled/auth/internal/queues/payloads"
	"github.com/roledio/roled/auth/internal/repositories"
	im "github.com/roledio/roled/auth/internal/repositories/interfaces/mocks"
	rm "github.com/roledio/roled/auth/internal/repositories/mocks"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	redismocks "github.com/roledio/roled/auth/pkg/redis/mocks"
	"github.com/shomali11/util/xhashes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateMemberPersistenceAndActivation(t *testing.T) {
	for _, failAt := range []string{"", "lookup", "role", "settings", "user", "user role", "member", "redis", "email"} {
		name := failAt
		if name == "" {
			name = "success"
		}
		t.Run(name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), constants.CtxAccount, &entities.Account{ID: "account", Name: "Account"})
			ctx = context.WithValue(ctx, constants.CtxAccessToken, &entities.AccessToken{ProjectID: "system"})
			reg := rm.NewMockRegistry(t)
			projects := im.NewMockProjectRepository(t)
			users := im.NewMockUserRepository(t)
			roles := im.NewMockRoleRepository(t)
			settings := im.NewMockProjectSettingRepository(t)
			userRoles := im.NewMockUserRoleRepository(t)
			members := im.NewMockMemberRepository(t)
			redis := redismocks.NewMockService(t)
			publisher := qm.NewMockPublisher(t)
			failure := errors.New("storage unavailable")
			reg.EXPECT().ProjectRepository().Return(projects).Once()
			projects.EXPECT().FindSystem(ctx).Return(&entities.Project{ID: "system", Name: "Console", IsActive: true, IsSystem: true}, nil).Once()
			var savedUser *entities.User
			var savedMember *entities.Member
			var activationKey string
			stages := []struct {
				name  string
				setup func(error)
			}{
				{"lookup", func(err error) {
					reg.EXPECT().UserRepository().Return(users).Once()
					users.EXPECT().FindByProjectIDAndEmail(ctx, "system", "member@example.com").Return(nil, err).Once()
				}},
				{"role", func(err error) {
					reg.EXPECT().RoleRepository().Return(roles).Once()
					roles.EXPECT().FindByProjectIDAndCode(ctx, "system", constants.RoledConsoleDefaultRoleCode).Return(&entities.Role{ID: "role"}, err).Once()
				}},
				{"settings", func(err error) {
					reg.EXPECT().Tx(mock.Anything).RunAndReturn(func(fn func(repositories.Registry) error) error { return fn(reg) }).Once()
					reg.EXPECT().ProjectSettingRepository().Return(settings).Once()
					settings.EXPECT().FindByProjectID(ctx, "system").Return(&entities.ProjectSetting{IsAllowTempEmail: true}, err).Once()
				}},
				{"user", func(err error) {
					reg.EXPECT().UserRepository().Return(users).Once()
					users.On("Create", ctx, mock.Anything).Run(func(a mock.Arguments) {
						savedUser = a.Get(1).(*entities.User)
						assert.Equal(t, "account", savedUser.AccountID)
						assert.Equal(t, "system", savedUser.ProjectID)
						assert.Equal(t, "member", savedUser.DisplayName)
						require.NotNil(t, savedUser.Email)
						assert.Equal(t, "member@example.com", *savedUser.Email)
						assert.False(t, savedUser.IsActive)
					}).Return(err).Once()
				}},
				{"user role", func(err error) {
					reg.EXPECT().UserRoleRepository().Return(userRoles).Once()
					userRoles.On("Create", ctx, mock.Anything).Run(func(a mock.Arguments) {
						assert.Equal(t, &entities.UserRole{UserID: savedUser.ID, RoleID: "role"}, a.Get(1))
					}).Return(err).Once()
				}},
				{"member", func(err error) {
					reg.EXPECT().MemberRepository().Return(members).Once()
					members.On("Create", ctx, mock.Anything).Run(func(a mock.Arguments) {
						savedMember = a.Get(1).(*entities.Member)
						assert.Equal(t, "account", savedMember.AccountID)
						assert.Equal(t, savedUser.ID, savedMember.UserID)
						assert.NotEmpty(t, savedMember.ID)
					}).Return(err).Once()
				}},
				{"redis", func(err error) {
					redis.On("SetData", ctx, mock.Anything, mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
						activationKey = a.String(1)
						assert.True(t, strings.HasPrefix(activationKey, rediskeys.ActivateMemberPrefix+":"))
						assert.Equal(t, models.CreateMemberTokenData{UserID: savedUser.ID}, a.Get(2))
						assert.Equal(t, time.Hour, a.Get(3))
					}).Return(err).Once()
				}},
				{"email", func(err error) {
					publisher.On("Publish", ctx, mock.Anything).Run(func(a mock.Arguments) {
						msg := a.Get(1).(queues.Message)
						var email payloads.EmailPayload
						require.NoError(t, json.Unmarshal([]byte(msg.Payload), &email))
						assert.Equal(t, "member@example.com", email.To)
						assert.Equal(t, "Account", email.AccountName)
						assert.Equal(t, "Console", email.ProjectName)
						assert.Equal(t, constants.EmailPayloadTypeActivateMember, email.Type)
						assert.Equal(t, "Console <no-reply@example.com>", email.From)
						assert.Len(t, email.Token, 64)
						assert.Equal(t, rediskeys.ActivateMemberPrefix+":"+xhashes.SHA256(email.Token), activationKey)
					}).Return(err).Once()
				}},
			}
			for _, stage := range stages {
				if stage.name == failAt {
					stage.setup(failure)
					break
				}
				stage.setup(nil)
			}
			cfg := &configs.DefaultConfig{ActivateMemberExpiryDuration: "1h"}
			cfg.Email.From = "%s <no-reply@example.com>"
			s := &memberService{registry: reg, defaultConfig: cfg, redisService: redis, emailPublisher: publisher}
			req := &models.CreateMemberRequest{AccountID: "another-account", Email: "Member@Example.com"}
			got, err := s.CreateMember(ctx, req)
			assert.Equal(t, "account", req.AccountID, "non-system account must not select another tenant")
			if failAt != "" {
				require.ErrorIs(t, err, pkgerrors.ErrSystemError)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, savedMember.ID, got.ID)
			assert.Equal(t, "Member@Example.com", got.Email)
			assert.Equal(t, "member", got.DisplayName)
			assert.False(t, got.IsActive)
		})
	}
}
