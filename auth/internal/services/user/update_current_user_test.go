package user

import (
	"context"
	"errors"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	im "github.com/roledio/roled/auth/internal/repositories/interfaces/mocks"
	rm "github.com/roledio/roled/auth/internal/repositories/mocks"
	"github.com/roledio/roled/auth/pkg/utils/passwordutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestUpdateCurrentUserIdentityAndTransaction(t *testing.T) {
	for _, stage := range []string{"success", "same email", "clear email", "password", "missing token", "non-user token", "find error", "not found", "duplicate email", "email lookup error", "write error", "no affected", "commit error"} {
		t.Run(stage, func(t *testing.T) {
			ctx := context.Background()
			reg, tx := rm.NewMockRegistry(t), rm.NewMockRegistry(t)
			users := im.NewMockUserRepository(t)
			failure := errors.New("offline")
			id := "user"
			oldEmail := "old@example.com"
			now := time.Now()
			oldHash := "old hash"
			value := &entities.User{ID: id, ProjectID: "project", Email: &oldEmail, EmailVerifiedAt: &now, PasswordHash: &oldHash, IsActive: true}
			req := &models.UpdateCurrentUserRequest{Email: " New@Example.COM ", DisplayName: "  New name  "}
			if stage == "same email" {
				req.Email = oldEmail
			}
			if stage == "clear email" {
				req.Email = ""
			}
			if stage == "password" {
				req.Password = "correct horse battery staple"
			}
			if stage != "missing token" {
				token := &entities.AccessToken{ProjectID: "project", UserID: &id}
				if stage == "non-user token" {
					token.UserID = nil
				}
				ctx = context.WithValue(ctx, constants.CtxAccessToken, token)
				if token.UserID != nil {
					reg.On("UserRepository").Return(users).Once()
					var e error
					if stage == "find error" {
						e = failure
					}
					if stage == "not found" {
						value = nil
					}
					users.On("FindByIDAndProjectID", ctx, id, "project").Return(value, e).Once()
					stop := e != nil || value == nil
					if !stop && stage != "same email" && stage != "clear email" {
						var existing *entities.User
						var e error
						if stage == "duplicate email" {
							existing = &entities.User{ID: "other"}
						}
						if stage == "email lookup error" {
							e = failure
						}
						users.On("FindByProjectIDAndEmail", ctx, "project", "new@example.com").Return(existing, e).Once()
						stop = e != nil || existing != nil
					}
					if !stop {
						tx.On("UserRepository").Return(users).Once()
						var e error
						if stage == "write error" {
							e = failure
						}
						n := 1
						if stage == "no affected" {
							n = 0
						}
						users.On("Update", ctx, mock.Anything).Run(func(a mock.Arguments) {
							u := a.Get(1).(*entities.User)
							require.Equal(t, "New name", u.DisplayName)
							require.Equal(t, id, u.ID)
							require.Equal(t, "project", u.ProjectID)
							switch stage {
							case "same email":
								require.Equal(t, &now, u.EmailVerifiedAt)
							case "clear email":
								require.Nil(t, u.Email)
							default:
								require.Equal(t, "new@example.com", *u.Email)
								require.Nil(t, u.EmailVerifiedAt)
							}
							if stage == "password" {
								require.True(t, passwordutil.IsValidPassword(req.Password, *u.PasswordHash))
								require.NotEqual(t, req.Password, *u.PasswordHash)
							} else {
								require.Equal(t, oldHash, *u.PasswordHash)
							}
						}).Return(n, e).Once()
						reg.On("Tx", mock.Anything).Return(func(fn func(repositories.Registry) error) error {
							if e := fn(tx); e != nil {
								return e
							}
							if stage == "commit error" {
								return failure
							}
							return nil
						}).Once()
					}
				}
			}
			result, err := (&userService{registry: reg}).UpdateCurrentUser(ctx, req)
			switch stage {
			case "success", "same email", "clear email", "password":
				require.NoError(t, err)
				require.Equal(t, "user", result.ID)
				require.Equal(t, "New name", result.DisplayName)
				require.True(t, result.IsActive)
			default:
				require.Error(t, err)
				require.Nil(t, result)
			}
		})
	}
}
