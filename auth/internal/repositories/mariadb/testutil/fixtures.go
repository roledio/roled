package testutil

import (
	"context"
	"fmt"
	"time"

	"github.com/roledio/roled/auth/internal/entities"
)

// AccountFixture represents test data for account
type AccountFixture struct {
	ID          string
	Name        string
	Description string
	IsActive    bool
	IsSystem    bool
}

// CreateAccount inserts an account into the database and returns the entity
func CreateAccount(ctx context.Context, db *TestDB, fixture AccountFixture) (*entities.Account, error) {
	now := time.Now().UTC().Truncate(time.Millisecond)

	account := &entities.Account{
		ID:          fixture.ID,
		Name:        fixture.Name,
		Description: fixture.Description,
		IsActive:    fixture.IsActive,
		IsSystem:    fixture.IsSystem,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	query := `INSERT INTO accounts (id, name, description, is_active, is_system, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := db.ExecContext(ctx, query,
		account.ID,
		account.Name,
		account.Description,
		account.IsActive,
		account.IsSystem,
		account.CreatedAt,
		account.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return account, nil
}

// CreateAccounts inserts multiple accounts into the database
func CreateAccounts(ctx context.Context, db *TestDB, fixtures []AccountFixture) ([]entities.Account, error) {
	accounts := make([]entities.Account, 0, len(fixtures))
	for _, f := range fixtures {
		account, err := CreateAccount(ctx, db, f)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, *account)
	}
	return accounts, nil
}

// DefaultAccountFixtures returns common test fixtures for accounts
func DefaultAccountFixtures() []AccountFixture {
	return []AccountFixture{
		{
			ID:          "acc_001_test_account_1",
			Name:        "Test Account 1",
			Description: "First test account",
			IsActive:    true,
			IsSystem:    false,
		},
		{
			ID:          "acc_002_test_account_2",
			Name:        "Test Account 2",
			Description: "Second test account",
			IsActive:    false,
			IsSystem:    false,
		},
		{
			ID:          "acc_003_system_account",
			Name:        "System Account",
			Description: "System managed account",
			IsActive:    true,
			IsSystem:    true,
		},
	}
}

// ProjectFixture represents test data for project
type ProjectFixture struct {
	ID          string
	AccountID   string
	Name        string
	Description *string
	LogoURL     *string
	IsActive    bool
	IsSystem    bool
}

// CreateProject inserts a project into the database and returns the entity
func CreateProject(ctx context.Context, db *TestDB, fixture ProjectFixture) (*entities.Project, error) {
	now := time.Now().UTC().Truncate(time.Millisecond)

	project := &entities.Project{
		ID:          fixture.ID,
		AccountID:   fixture.AccountID,
		Name:        fixture.Name,
		Description: fixture.Description,
		LogoURL:     fixture.LogoURL,
		IsActive:    fixture.IsActive,
		IsSystem:    fixture.IsSystem,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	query := `INSERT INTO projects (id, account_id, name, description, logo_url, is_active, is_system, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.ExecContext(ctx, query,
		project.ID,
		project.AccountID,
		project.Name,
		project.Description,
		project.LogoURL,
		project.IsActive,
		project.IsSystem,
		project.CreatedAt,
		project.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return project, nil
}

// ClientFixture represents test data for client
type ClientFixture struct {
	ID              string
	AccountID       string
	ProjectID       string
	Name            string
	Description     *string
	SecretEncrypted string
	IsActive        bool
	IsDefault       bool
}

// CreateClient inserts a client into the database and returns the entity
func CreateClient(ctx context.Context, db *TestDB, fixture ClientFixture) (*entities.Client, error) {
	now := time.Now().UTC().Truncate(time.Millisecond)

	client := &entities.Client{
		ID:              fixture.ID,
		AccountID:       fixture.AccountID,
		ProjectID:       fixture.ProjectID,
		Name:            fixture.Name,
		Description:     fixture.Description,
		SecretEncrypted: fixture.SecretEncrypted,
		IsActive:        fixture.IsActive,
		IsDefault:       fixture.IsDefault,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	query := `INSERT INTO clients (id, account_id, project_id, name, description, secret_encrypted, is_active, is_default, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.ExecContext(ctx, query,
		client.ID,
		client.AccountID,
		client.ProjectID,
		client.Name,
		client.Description,
		client.SecretEncrypted,
		client.IsActive,
		client.IsDefault,
		client.CreatedAt,
		client.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return client, nil
}

// UserFixture represents test data for user
type UserFixture struct {
	ID          string
	AccountID   string
	ProjectID   string
	DisplayName string
	Email       *string
	AvatarURL   *string
	ExternalID  *string
	IsActive    bool
}

// CreateUser inserts a user into the database and returns the entity
func CreateUser(ctx context.Context, db *TestDB, fixture UserFixture) (*entities.User, error) {
	now := time.Now().UTC().Truncate(time.Millisecond)

	user := &entities.User{
		ID:             fixture.ID,
		AccountID:      fixture.AccountID,
		ProjectID:      fixture.ProjectID,
		DisplayName:    fixture.DisplayName,
		Email:          fixture.Email,
		AvatarURL:      fixture.AvatarURL,
		ExternalUserID: fixture.ExternalID,
		IsActive:       fixture.IsActive,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	query := `INSERT INTO users (id, account_id, project_id, display_name, email, avatar_url, external_user_id, is_active, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.ExecContext(ctx, query,
		user.ID,
		user.AccountID,
		user.ProjectID,
		user.DisplayName,
		user.Email,
		user.AvatarURL,
		user.ExternalUserID,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// RoleFixture represents test data for role
type RoleFixture struct {
	ID          string
	AccountID   string
	ProjectID   string
	Code        string
	Name        string
	Description string
}

// CreateRole inserts a role into the database and returns the entity
func CreateRole(ctx context.Context, db *TestDB, fixture RoleFixture) (*entities.Role, error) {
	now := time.Now().UTC().Truncate(time.Millisecond)

	role := &entities.Role{
		ID:          fixture.ID,
		AccountID:   fixture.AccountID,
		ProjectID:   fixture.ProjectID,
		Code:        fixture.Code,
		Name:        fixture.Name,
		Description: fixture.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	query := `INSERT INTO roles (id, account_id, project_id, code, name, description, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.ExecContext(ctx, query,
		role.ID,
		role.AccountID,
		role.ProjectID,
		role.Code,
		role.Name,
		role.Description,
		role.CreatedAt,
		role.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return role, nil
}

// AccessTokenFixture represents test data for access token
type AccessTokenFixture struct {
	ID        string
	AccountID string
	ProjectID string
	ClientID  string
	UserID    *string
	GrantType string
	Status    string
	ExpiresIn int
	IssuedAt  time.Time
}

// CreateAccessToken inserts an access token into the database and returns the entity
func CreateAccessToken(ctx context.Context, db *TestDB, fixture AccessTokenFixture) (*entities.AccessToken, error) {
	now := time.Now().UTC().Truncate(time.Millisecond)

	token := &entities.AccessToken{
		ID:        fixture.ID,
		AccountID: fixture.AccountID,
		ProjectID: fixture.ProjectID,
		ClientID:  fixture.ClientID,
		UserID:    fixture.UserID,
		GrantType: fixture.GrantType,
		Status:    fixture.Status,
		ExpiresIn: &fixture.ExpiresIn,
		IssuedAt:  &fixture.IssuedAt,
		CreatedAt: now,
		UpdatedAt: now,
	}

	query := `INSERT INTO access_tokens (id, account_id, project_id, client_id, user_id, grant_type, status, expires_in, issued_at, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.ExecContext(ctx, query,
		token.ID,
		token.AccountID,
		token.ProjectID,
		token.ClientID,
		token.UserID,
		token.GrantType,
		token.Status,
		token.ExpiresIn,
		token.IssuedAt,
		token.CreatedAt,
		token.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	return token, nil
}
