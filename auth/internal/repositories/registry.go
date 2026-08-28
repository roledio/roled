package repositories

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofiber/fiber/v3/log"
	"github.com/jmoiron/sqlx"
	"github.com/roledio/roled/internal/configs"
	"github.com/roledio/roled/internal/repositories/interfaces"
	"github.com/roledio/roled/internal/repositories/mariadb"
	"github.com/roledio/roled/internal/repositories/redis"
	"github.com/roledio/roled/internal/services/infra"
)

type Registry interface {
	Ping() error
	Tx(fn func(registry Registry) error) error
	AccountRepository() interfaces.AccountRepository
	MemberRepository() interfaces.MemberRepository
	ProjectRepository() interfaces.ProjectRepository
	ProjectSettingRepository() interfaces.ProjectSettingRepository
	RedirectURIRepository() interfaces.RedirectURIRepository
	RoleRepository() interfaces.RoleRepository
	AuthCodeRepository() interfaces.AuthCodeRepository
	AccessTokenRepository() interfaces.AccessTokenRepository
	RefreshTokenRepository() interfaces.RefreshTokenRepository
	UserRepository() interfaces.UserRepository
	UserRoleRepository() interfaces.UserRoleRepository
	ResourceRepository() interfaces.ResourceRepository
	PermissionRepository() interfaces.PermissionRepository
	RolePermissionRepository() interfaces.RolePermissionRepository
	ClientRepository() interfaces.ClientRepository
	ClientPermissionRepository() interfaces.ClientPermissionRepository
}

type registry struct {
	defaultConfig *configs.DefaultConfig
	qx            interfaces.QueryExecutor
	redisService  infra.RedisService
}

func NewRegistry(defaultConfig *configs.DefaultConfig, db *sqlx.DB, redis ...infra.RedisService) Registry {
	sq.StatementBuilder = sq.StatementBuilder.PlaceholderFormat(sq.Question)
	reg := registry{
		defaultConfig: defaultConfig,
		qx:            db,
	}
	if len(redis) > 0 {
		reg.redisService = redis[0]
	}
	return &reg
}

func (r *registry) Ping() error {
	db, ok := r.qx.(*sqlx.DB)
	if !ok {
		return fmt.Errorf("cannot ping: not a database connection")
	}
	// Ping the database with a timeout of 5s
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return db.PingContext(ctx)
}

func (c *registry) Tx(fn func(registry Registry) error) error {
	db, ok := c.qx.(*sqlx.DB)
	if !ok {
		return fmt.Errorf("nested transaction not supported")
	}
	tx, err := db.BeginTxx(context.Background(), nil)
	if err != nil {
		return err
	}
	var errtx error
	defer func() {
		if errtx != nil {
			if err := tx.Rollback(); err != nil {
				log.Errorw("Failed to rollback transaction", "error", err)
			}
		}
	}()
	errtx = fn(&registry{
		defaultConfig: c.defaultConfig,
		qx:            tx,
		redisService:  c.redisService,
	})
	if errtx != nil {
		return errtx
	}
	errtx = tx.Commit()
	return errtx
}

func (r *registry) AccountRepository() interfaces.AccountRepository {
	repo := mariadb.NewAccountRepository(r.qx)
	if r.redisService != nil {
		return redis.NewAccountRepository(repo,
			r.redisService,
			r.defaultConfig.CacheDefaultTTLDuration)
	}
	return repo
}

func (r *registry) MemberRepository() interfaces.MemberRepository {
	repo := mariadb.NewMemberRepository(r.qx)
	if r.redisService != nil {
		return redis.NewMemberRepository(repo,
			r.redisService,
			r.defaultConfig.CacheDefaultTTLDuration)
	}
	return repo
}

func (r *registry) ProjectRepository() interfaces.ProjectRepository {
	repo := mariadb.NewProjectRepository(r.qx)
	if r.redisService != nil {
		return redis.NewProjectCacheRepository(repo,
			r.redisService,
			r.defaultConfig.CacheDefaultTTLDuration)
	}
	return repo
}

func (r *registry) ProjectSettingRepository() interfaces.ProjectSettingRepository {
	repo := mariadb.NewProjectSettingRepository(r.qx)
	if r.redisService != nil {
		return redis.NewProjectSettingRepository(repo,
			r.redisService,
			r.defaultConfig.CacheDefaultTTLDuration)
	}
	return repo
}

func (r *registry) RedirectURIRepository() interfaces.RedirectURIRepository {
	repo := mariadb.NewRedirectURIRepository(r.qx)
	if r.redisService != nil {
		return redis.NewRedirectURIRepository(repo,
			r.redisService,
			r.defaultConfig.CacheDefaultTTLDuration)
	}
	return repo
}

func (r *registry) RoleRepository() interfaces.RoleRepository {
	repo := mariadb.NewRoleRepository(r.qx)
	if r.redisService != nil {
		return redis.NewRoleRepository(repo,
			r.redisService,
			r.defaultConfig.CacheDefaultTTLDuration)
	}
	return repo
}

func (r *registry) AuthCodeRepository() interfaces.AuthCodeRepository {
	repo := mariadb.NewAuthCodeRepository(r.qx)
	if r.redisService != nil {
		return redis.NewAuthCodeRepository(repo,
			r.redisService,
			r.defaultConfig.CacheDefaultTTLDuration)
	}
	return repo
}

func (r *registry) AccessTokenRepository() interfaces.AccessTokenRepository {
	repo := mariadb.NewAccessTokenRepository(r.qx)
	if r.redisService != nil {
		return redis.NewAccessTokenRepository(repo,
			r.redisService,
			r.defaultConfig.CacheDefaultTTLDuration)
	}
	return repo
}

func (r *registry) RefreshTokenRepository() interfaces.RefreshTokenRepository {
	repo := mariadb.NewRefreshTokenRepository(r.qx)
	if r.redisService != nil {
		return redis.NewRefreshTokenRepository(repo,
			r.redisService,
			r.defaultConfig.CacheDefaultTTLDuration)
	}
	return repo
}

func (r *registry) UserRepository() interfaces.UserRepository {
	repo := mariadb.NewUserRepository(r.qx)
	if r.redisService != nil {
		return redis.NewUserRepository(repo,
			r.redisService,
			r.defaultConfig.CacheDefaultTTLDuration)
	}
	return repo
}

func (r *registry) UserRoleRepository() interfaces.UserRoleRepository {
	return mariadb.NewUserRoleRepository(r.qx)
}

func (r *registry) ResourceRepository() interfaces.ResourceRepository {
	repo := mariadb.NewResourceRepository(r.qx)
	if r.redisService != nil {
		return redis.NewResourceRepository(repo,
			r.redisService,
			r.defaultConfig.CacheDefaultTTLDuration)
	}
	return repo
}

func (r *registry) PermissionRepository() interfaces.PermissionRepository {
	repo := mariadb.NewPermissionRepository(r.qx)
	if r.redisService != nil {
		return redis.NewPermissionRepository(repo,
			r.redisService,
			r.defaultConfig.CacheDefaultTTLDuration)
	}
	return repo
}

func (r *registry) RolePermissionRepository() interfaces.RolePermissionRepository {
	return mariadb.NewRolePermissionRepository(r.qx)
}

func (r *registry) ClientRepository() interfaces.ClientRepository {
	repo := mariadb.NewClientRepository(r.qx)
	if r.redisService != nil {
		return redis.NewClientRepository(repo,
			r.redisService,
			r.defaultConfig.CacheDefaultTTLDuration)
	}
	return repo
}

func (r *registry) ClientPermissionRepository() interfaces.ClientPermissionRepository {
	return mariadb.NewClientPermissionRepository(r.qx)
}
