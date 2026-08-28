package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/pressly/goose/v3"
	"github.com/roledio/roled/pkg/constants"
	"github.com/roledio/roled/pkg/utils/encryptionutil"
	"github.com/roledio/roled/pkg/utils/idutil"
	"github.com/roledio/roled/pkg/utils/passwordutil"
)

func init() {
	goose.AddMigrationContext(up000002, down000002)
}

func up000002(ctx context.Context, tx *sql.Tx) error {
	// Generate all IDs upfront
	ids := generateSeedIDs()

	// Seed data in order
	if err := seedAccounts(ctx, tx, ids); err != nil {
		return err
	}
	if err := seedProjects(ctx, tx, ids); err != nil {
		return err
	}
	if err := seedRedirectURIs(ctx, tx, ids); err != nil {
		return err
	}
	if err := seedProjectSettings(ctx, tx, ids); err != nil {
		return err
	}
	if err := seedResources(ctx, tx, ids); err != nil {
		return err
	}
	if err := seedPermissions(ctx, tx, ids); err != nil {
		return err
	}
	if err := seedRoles(ctx, tx, ids); err != nil {
		return err
	}
	if err := seedRolePermissions(ctx, tx, ids); err != nil {
		return err
	}
	if err := seedUsers(ctx, tx, ids); err != nil {
		return err
	}
	if err := seedMembers(ctx, tx, ids); err != nil {
		return err
	}
	if err := seedUserRoles(ctx, tx, ids); err != nil {
		return err
	}
	if err := seedClients(ctx, tx, ids); err != nil {
		return err
	}
	if err := seedClientPermissions(ctx, tx, ids); err != nil {
		return err
	}
	// Set flag to indicate that initial data has been generated
	_ = os.Setenv("INITIAL_DATA_GENERATED", "true")
	return nil
}

func down000002(ctx context.Context, tx *sql.Tx) error {
	return nil
}

// seedIDs holds all generated IDs for seed data
type seedIDs struct {
	SystemAccountID       string
	SystemAccountMemberID string
	UserAccountID         string
	UserAccountMemberID   string
	SystemProjectID       string
	ProjectSettingID      string
	DefaultRoleID         string
	SystemUserID          string
	UserUserID            string
	ClientID              string
	AccountsResourceID    string
	MembersResourceID     string
	ProjectsResourceID    string
	ResourcesResourceID   string
	PermissionsResourceID string
	UsersResourceID       string
	RolesResourceID       string
	ClientsResourceID     string
	PermissionIDs         map[string]string
}

func generateSeedIDs() *seedIDs {
	ids := &seedIDs{
		SystemAccountID:       idutil.NewID(),
		SystemAccountMemberID: idutil.NewID(),
		UserAccountID:         idutil.NewID(),
		UserAccountMemberID:   idutil.NewID(),
		SystemProjectID:       idutil.NewID(),
		ProjectSettingID:      idutil.NewID(),
		DefaultRoleID:         idutil.NewID(),
		SystemUserID:          idutil.NewID(),
		UserUserID:            idutil.NewID(),
		ClientID:              idutil.NewID(),
		AccountsResourceID:    idutil.NewID(),
		MembersResourceID:     idutil.NewID(),
		ProjectsResourceID:    idutil.NewID(),
		ResourcesResourceID:   idutil.NewID(),
		PermissionsResourceID: idutil.NewID(),
		UsersResourceID:       idutil.NewID(),
		RolesResourceID:       idutil.NewID(),
		ClientsResourceID:     idutil.NewID(),
		PermissionIDs:         make(map[string]string),
	}

	// Generate permission IDs
	permissionNames := []string{
		"read_accounts", "create_accounts", "update_accounts", "delete_accounts",
		"read_members", "create_members", "update_members", "delete_members",
		"read_projects", "create_projects", "update_projects", "delete_projects",
		"read_resources", "create_resources", "update_resources", "delete_resources",
		"read_permissions", "create_permissions", "update_permissions", "delete_permissions",
		"read_users", "create_users", "update_users", "delete_users",
		"read_roles", "create_roles", "update_roles", "delete_roles",
		"read_clients", "create_clients", "update_clients", "delete_clients",
	}
	for _, name := range permissionNames {
		ids.PermissionIDs[name] = idutil.NewID()
	}

	return ids
}

func seedAccounts(ctx context.Context, tx *sql.Tx, ids *seedIDs) error {
	accounts := []struct {
		ID          string
		Name        string
		Description string
		IsActive    bool
		IsSystem    bool
	}{
		{ids.SystemAccountID, "Roled (System)", "Roled system account", true, true},
		{ids.UserAccountID, "Roled (User)", "Roled user account", true, false},
	}
	for _, acc := range accounts {
		_, err := tx.ExecContext(ctx, `INSERT INTO accounts (id, name, description, is_active, is_system) VALUES (?, ?, ?, ?, ?)`, acc.ID, acc.Name, acc.Description, acc.IsActive, acc.IsSystem)
		if err != nil {
			return err
		}
	}
	return nil
}

func seedMembers(ctx context.Context, tx *sql.Tx, ids *seedIDs) error {
	members := []struct {
		ID        string
		AccountID string
		UserID    string
		IsAdmin   bool
	}{
		{ids.SystemAccountMemberID, ids.SystemAccountID, ids.SystemUserID, true},
		{ids.UserAccountMemberID, ids.UserAccountID, ids.UserUserID, true},
	}
	for _, mem := range members {
		_, err := tx.ExecContext(ctx, `INSERT INTO members (id, account_id, user_id, is_admin) VALUES (?, ?, ?, ?)`, mem.ID, mem.AccountID, mem.UserID, mem.IsAdmin)
		if err != nil {
			return err
		}
	}
	return nil
}

func seedProjects(ctx context.Context, tx *sql.Tx, ids *seedIDs) error {
	name := "Roled Console"
	description := "Roled Console is a web-based control panel for managing projects, users, and roles using Roled's centralized authentication and authorization platform."
	logoURL := "https://static.roled.io/file/roled-production/project-logo/roled-logo.png"
	projects := []struct {
		ID          string
		AccountID   string
		Name        string
		Description string
		LogoURL     string
		IsActive    bool
		IsSystem    bool
	}{
		{ids.SystemProjectID, ids.SystemAccountID, name, description, logoURL, true, true},
	}
	for _, proj := range projects {
		_, err := tx.ExecContext(ctx, `INSERT INTO projects (id, account_id, name, description, logo_url, is_active, is_system) VALUES (?, ?, ?, ?, ?, ?, ?)`, proj.ID, proj.AccountID, proj.Name, proj.Description, proj.LogoURL, proj.IsActive, proj.IsSystem)
		if err != nil {
			return err
		}
	}
	return nil
}

func seedRedirectURIs(ctx context.Context, tx *sql.Tx, ids *seedIDs) error {
	redirectURIs := []struct {
		ProjectID   string
		RedirectURI string
		LoginURL    string
	}{
		{ids.SystemProjectID, "http://localhost:4000/signin/callback", "http://localhost:4000/signin"},
		{ids.SystemProjectID, "https://console-staging.roled.io/signin/callback", "https://console-staging.roled.io/signin"},
		{ids.SystemProjectID, "https://console.roled.io/signin/callback", "https://console.roled.io/signin"},
	}
	for _, uri := range redirectURIs {
		_, err := tx.ExecContext(ctx, `INSERT INTO redirect_uris (project_id, redirect_uri, login_url) VALUES (?, ?, ?)`, uri.ProjectID, uri.RedirectURI, uri.LoginURL)
		if err != nil {
			return err
		}
	}
	return nil
}

func seedProjectSettings(ctx context.Context, tx *sql.Tx, ids *seedIDs) error {
	projectSettings := []struct {
		ID                      string
		ProjectID               string
		IsSignupEnabled         bool
		IsSignupVerifyEmail     bool
		IsForgotPasswordEnabled bool
		IsAllowTempEmail        bool
	}{
		{ids.ProjectSettingID, ids.SystemProjectID, true, true, true, true},
	}
	for _, ps := range projectSettings {
		_, err := tx.ExecContext(ctx, `INSERT INTO project_settings (id, project_id, is_signup_enabled, is_signup_verify_email, is_forgot_password_enabled, is_allow_temp_email) VALUES (?, ?, ?, ?, ?, ?)`, ps.ID, ps.ProjectID, ps.IsSignupEnabled, ps.IsSignupVerifyEmail, ps.IsForgotPasswordEnabled, ps.IsAllowTempEmail)
		if err != nil {
			return err
		}
	}
	return nil
}

func seedResources(ctx context.Context, tx *sql.Tx, ids *seedIDs) error {
	resources := []struct {
		ID          string
		AccountID   string
		ProjectID   string
		Code        string
		Name        string
		Description string
	}{
		{ids.AccountsResourceID, ids.SystemAccountID, ids.SystemProjectID, "accounts", "Accounts", "Accounts resource"},
		{ids.MembersResourceID, ids.SystemAccountID, ids.SystemProjectID, "members", "Members", "Members resource"},
		{ids.ProjectsResourceID, ids.SystemAccountID, ids.SystemProjectID, "projects", "Projects", "Projects resource"},
		{ids.ResourcesResourceID, ids.SystemAccountID, ids.SystemProjectID, "resources", "Resources", "Resources resource"},
		{ids.PermissionsResourceID, ids.SystemAccountID, ids.SystemProjectID, "permissions", "Permissions", "Permissions resource"},
		{ids.UsersResourceID, ids.SystemAccountID, ids.SystemProjectID, "users", "Users", "Users resource"},
		{ids.RolesResourceID, ids.SystemAccountID, ids.SystemProjectID, "roles", "Roles", "Roles resource"},
		{ids.ClientsResourceID, ids.SystemAccountID, ids.SystemProjectID, "clients", "Clients", "Clients resource"},
	}
	for _, res := range resources {
		_, err := tx.ExecContext(ctx, `INSERT INTO resources (id, account_id, project_id, code, name, description, is_default) VALUES (?, ?, ?, ?, ?, ?, ?)`, res.ID, res.AccountID, res.ProjectID, res.Code, res.Name, res.Description, true)
		if err != nil {
			return err
		}
	}
	return nil
}

func seedPermissions(ctx context.Context, tx *sql.Tx, ids *seedIDs) error {
	permissions := []struct {
		ID          string
		ResourceID  string
		Code        string
		Name        string
		Description string
	}{
		{ids.PermissionIDs["read_accounts"], ids.AccountsResourceID, "read", "Read", "Read accounts"},
		{ids.PermissionIDs["create_accounts"], ids.AccountsResourceID, "create", "Create", "Create accounts"},
		{ids.PermissionIDs["update_accounts"], ids.AccountsResourceID, "update", "Update", "Update accounts"},
		{ids.PermissionIDs["delete_accounts"], ids.AccountsResourceID, "delete", "Delete", "Delete accounts"},
		{ids.PermissionIDs["read_members"], ids.MembersResourceID, "read", "Read", "Read members"},
		{ids.PermissionIDs["create_members"], ids.MembersResourceID, "create", "Create", "Create members"},
		{ids.PermissionIDs["update_members"], ids.MembersResourceID, "update", "Update", "Update members"},
		{ids.PermissionIDs["delete_members"], ids.MembersResourceID, "delete", "Delete", "Delete members"},
		{ids.PermissionIDs["read_projects"], ids.ProjectsResourceID, "read", "Read", "Read projects"},
		{ids.PermissionIDs["create_projects"], ids.ProjectsResourceID, "create", "Create", "Create projects"},
		{ids.PermissionIDs["update_projects"], ids.ProjectsResourceID, "update", "Update", "Update projects"},
		{ids.PermissionIDs["delete_projects"], ids.ProjectsResourceID, "delete", "Delete", "Delete projects"},
		{ids.PermissionIDs["read_resources"], ids.ResourcesResourceID, "read", "Read", "Read resources"},
		{ids.PermissionIDs["create_resources"], ids.ResourcesResourceID, "create", "Create", "Create resources"},
		{ids.PermissionIDs["update_resources"], ids.ResourcesResourceID, "update", "Update", "Update resources"},
		{ids.PermissionIDs["delete_resources"], ids.ResourcesResourceID, "delete", "Delete", "Delete resources"},
		{ids.PermissionIDs["read_permissions"], ids.PermissionsResourceID, "read", "Read", "Read permissions"},
		{ids.PermissionIDs["create_permissions"], ids.PermissionsResourceID, "create", "Create", "Create permissions"},
		{ids.PermissionIDs["update_permissions"], ids.PermissionsResourceID, "update", "Update", "Update permissions"},
		{ids.PermissionIDs["delete_permissions"], ids.PermissionsResourceID, "delete", "Delete", "Delete permissions"},
		{ids.PermissionIDs["read_users"], ids.UsersResourceID, "read", "Read", "Read users"},
		{ids.PermissionIDs["create_users"], ids.UsersResourceID, "create", "Create", "Create users"},
		{ids.PermissionIDs["update_users"], ids.UsersResourceID, "update", "Update", "Update users"},
		{ids.PermissionIDs["delete_users"], ids.UsersResourceID, "delete", "Delete", "Delete users"},
		{ids.PermissionIDs["read_roles"], ids.RolesResourceID, "read", "Read", "Read roles"},
		{ids.PermissionIDs["create_roles"], ids.RolesResourceID, "create", "Create", "Create roles"},
		{ids.PermissionIDs["update_roles"], ids.RolesResourceID, "update", "Update", "Update roles"},
		{ids.PermissionIDs["delete_roles"], ids.RolesResourceID, "delete", "Delete", "Delete roles"},
		{ids.PermissionIDs["read_clients"], ids.ClientsResourceID, "read", "Read", "Read clients"},
		{ids.PermissionIDs["create_clients"], ids.ClientsResourceID, "create", "Create", "Create clients"},
		{ids.PermissionIDs["update_clients"], ids.ClientsResourceID, "update", "Update", "Update clients"},
		{ids.PermissionIDs["delete_clients"], ids.ClientsResourceID, "delete", "Delete", "Delete clients"},
	}
	for _, perm := range permissions {
		_, err := tx.ExecContext(ctx, `INSERT INTO permissions (id, resource_id, code, name, description, is_default) VALUES (?, ?, ?, ?, ?, ?)`, perm.ID, perm.ResourceID, perm.Code, perm.Name, perm.Description, true)
		if err != nil {
			return err
		}
	}
	return nil
}

func seedRoles(ctx context.Context, tx *sql.Tx, ids *seedIDs) error {
	roles := []struct {
		ID          string
		AccountID   string
		ProjectID   string
		Code        string
		Name        string
		Description string
	}{
		{ids.DefaultRoleID, ids.SystemAccountID, ids.SystemProjectID, "default", "Default", "Default and the only role for Roled Console"},
	}
	for _, role := range roles {
		_, err := tx.ExecContext(ctx, `INSERT INTO roles (id, account_id, project_id, code, name, description) VALUES (?, ?, ?, ?, ?, ?)`, role.ID, role.AccountID, role.ProjectID, role.Code, role.Name, role.Description)
		if err != nil {
			return err
		}
	}
	// Set default role ID to project settings
	_, err := tx.ExecContext(ctx, `UPDATE project_settings SET default_signup_role_id = ? WHERE id = ?`, ids.DefaultRoleID, ids.ProjectSettingID)
	if err != nil {
		return err
	}
	return nil
}

func seedRolePermissions(ctx context.Context, tx *sql.Tx, ids *seedIDs) error {
	rolePermissions := []struct {
		RoleID       string
		PermissionID string
	}{
		{ids.DefaultRoleID, ids.PermissionIDs["read_accounts"]},
		{ids.DefaultRoleID, ids.PermissionIDs["update_accounts"]},
		{ids.DefaultRoleID, ids.PermissionIDs["delete_accounts"]},
		{ids.DefaultRoleID, ids.PermissionIDs["read_members"]},
		{ids.DefaultRoleID, ids.PermissionIDs["create_members"]},
		{ids.DefaultRoleID, ids.PermissionIDs["update_members"]},
		{ids.DefaultRoleID, ids.PermissionIDs["delete_members"]},
		{ids.DefaultRoleID, ids.PermissionIDs["read_projects"]},
		{ids.DefaultRoleID, ids.PermissionIDs["create_projects"]},
		{ids.DefaultRoleID, ids.PermissionIDs["update_projects"]},
		{ids.DefaultRoleID, ids.PermissionIDs["delete_projects"]},
		{ids.DefaultRoleID, ids.PermissionIDs["read_resources"]},
		{ids.DefaultRoleID, ids.PermissionIDs["create_resources"]},
		{ids.DefaultRoleID, ids.PermissionIDs["update_resources"]},
		{ids.DefaultRoleID, ids.PermissionIDs["delete_resources"]},
		{ids.DefaultRoleID, ids.PermissionIDs["read_permissions"]},
		{ids.DefaultRoleID, ids.PermissionIDs["create_permissions"]},
		{ids.DefaultRoleID, ids.PermissionIDs["update_permissions"]},
		{ids.DefaultRoleID, ids.PermissionIDs["delete_permissions"]},
		{ids.DefaultRoleID, ids.PermissionIDs["read_users"]},
		{ids.DefaultRoleID, ids.PermissionIDs["create_users"]},
		{ids.DefaultRoleID, ids.PermissionIDs["update_users"]},
		{ids.DefaultRoleID, ids.PermissionIDs["delete_users"]},
		{ids.DefaultRoleID, ids.PermissionIDs["read_roles"]},
		{ids.DefaultRoleID, ids.PermissionIDs["create_roles"]},
		{ids.DefaultRoleID, ids.PermissionIDs["update_roles"]},
		{ids.DefaultRoleID, ids.PermissionIDs["delete_roles"]},
		{ids.DefaultRoleID, ids.PermissionIDs["read_clients"]},
		{ids.DefaultRoleID, ids.PermissionIDs["create_clients"]},
		{ids.DefaultRoleID, ids.PermissionIDs["update_clients"]},
		{ids.DefaultRoleID, ids.PermissionIDs["delete_clients"]},
	}
	for _, rp := range rolePermissions {
		_, err := tx.ExecContext(ctx, `INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)`, rp.RoleID, rp.PermissionID)
		if err != nil {
			return err
		}
	}
	return nil
}

func seedUsers(ctx context.Context, tx *sql.Tx, ids *seedIDs) error {
	adminEmail := "admin@roled.io"
	adminPassword := idutil.NanoID(16)
	adminPasswordHash, err := passwordutil.HashPassword(adminPassword)
	if err != nil {
		return err
	}
	userEmail := "user@roled.io"
	userPassword := idutil.NanoID(16)
	userPasswordHash, err := passwordutil.HashPassword(userPassword)
	if err != nil {
		return err
	}
	now := time.Now()
	users := []struct {
		ID              string
		AccountID       string
		ProjectID       string
		Email           string
		PasswordHash    string
		DisplayName     string
		IsActive        bool
		EmailVerifiedAt *time.Time
	}{
		{ids.SystemUserID, ids.SystemAccountID, ids.SystemProjectID, adminEmail, adminPasswordHash, "Roled Admin", true, &now},
		{ids.UserUserID, ids.UserAccountID, ids.SystemProjectID, userEmail, userPasswordHash, "Roled User", true, &now},
	}
	for _, user := range users {
		_, err := tx.ExecContext(ctx, `INSERT INTO users (id, account_id, project_id, email, password_hash, display_name, is_active, email_verified_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, user.ID, user.AccountID, user.ProjectID, user.Email, user.PasswordHash, user.DisplayName, user.IsActive, user.EmailVerifiedAt)
		if err != nil {
			return err
		}
	}
	// Set the generated client id and secret as environment variables for later use
	_ = os.Setenv("GENERATED_ADMIN_EMAIL", adminEmail)
	_ = os.Setenv("GENERATED_ADMIN_PASSWORD", adminPassword)
	_ = os.Setenv("GENERATED_USER_EMAIL", userEmail)
	_ = os.Setenv("GENERATED_USER_PASSWORD", userPassword)
	return nil
}

func seedUserRoles(ctx context.Context, tx *sql.Tx, ids *seedIDs) error {
	userRoles := []struct {
		UserID string
		RoleID string
	}{
		{ids.SystemUserID, ids.DefaultRoleID},
		{ids.UserUserID, ids.DefaultRoleID},
	}
	for _, ur := range userRoles {
		_, err := tx.ExecContext(ctx, `INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)`, ur.UserID, ur.RoleID)
		if err != nil {
			return err
		}
	}
	return nil
}

func seedClients(ctx context.Context, tx *sql.Tx, ids *seedIDs) error {
	encryptionMasterKey := os.Getenv("ENCRYPTION_MASTER_KEY")
	if encryptionMasterKey == "" {
		return fmt.Errorf("environment variable ENCRYPTION_MASTER_KEY is not set")
	}
	purpose := constants.KeyPurposeClientSecret
	derivedKey, err := encryptionutil.DeriveKey([]byte(encryptionMasterKey), purpose)
	if err != nil {
		return err
	}
	secret := idutil.NanoID(64)
	secretEncrypted, err := encryptionutil.EncryptAES(secret, derivedKey, purpose)
	if err != nil {
		return err
	}
	name := "Main Client"
	description := "The main client for Roled Console"
	clients := []struct {
		ID              string
		AccountID       string
		ProjectID       string
		Name            string
		Description     string
		SecretEncrypted string
		IsActive        bool
		IsDefault       bool
	}{
		{ids.ClientID, ids.SystemAccountID, ids.SystemProjectID, name, description, secretEncrypted, true, true},
	}
	for _, client := range clients {
		_, err := tx.ExecContext(ctx, `INSERT INTO clients (id, account_id, project_id, name, description, secret_encrypted, is_active, is_default) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, client.ID, client.AccountID, client.ProjectID, client.Name, client.Description, client.SecretEncrypted, client.IsActive, client.IsDefault)
		if err != nil {
			return err
		}
	}
	// Set the generated client id and secret as environment variables for later use
	_ = os.Setenv("GENERATED_CLIENT_ID", ids.ClientID)
	_ = os.Setenv("GENERATED_CLIENT_SECRET", secret)
	return nil
}

func seedClientPermissions(ctx context.Context, tx *sql.Tx, ids *seedIDs) error {
	clientPermissions := []struct {
		ClientID     string
		PermissionID string
	}{
		{ids.ClientID, ids.PermissionIDs["read_accounts"]},
		{ids.ClientID, ids.PermissionIDs["create_accounts"]},
		{ids.ClientID, ids.PermissionIDs["update_accounts"]},
		{ids.ClientID, ids.PermissionIDs["delete_accounts"]},
		{ids.ClientID, ids.PermissionIDs["read_members"]},
		{ids.ClientID, ids.PermissionIDs["create_members"]},
		{ids.ClientID, ids.PermissionIDs["update_members"]},
		{ids.ClientID, ids.PermissionIDs["delete_members"]},
		{ids.ClientID, ids.PermissionIDs["read_projects"]},
		{ids.ClientID, ids.PermissionIDs["create_projects"]},
		{ids.ClientID, ids.PermissionIDs["update_projects"]},
		{ids.ClientID, ids.PermissionIDs["delete_projects"]},
		{ids.ClientID, ids.PermissionIDs["read_resources"]},
		{ids.ClientID, ids.PermissionIDs["create_resources"]},
		{ids.ClientID, ids.PermissionIDs["update_resources"]},
		{ids.ClientID, ids.PermissionIDs["delete_resources"]},
		{ids.ClientID, ids.PermissionIDs["read_permissions"]},
		{ids.ClientID, ids.PermissionIDs["create_permissions"]},
		{ids.ClientID, ids.PermissionIDs["update_permissions"]},
		{ids.ClientID, ids.PermissionIDs["delete_permissions"]},
		{ids.ClientID, ids.PermissionIDs["read_users"]},
		{ids.ClientID, ids.PermissionIDs["create_users"]},
		{ids.ClientID, ids.PermissionIDs["update_users"]},
		{ids.ClientID, ids.PermissionIDs["delete_users"]},
		{ids.ClientID, ids.PermissionIDs["read_roles"]},
		{ids.ClientID, ids.PermissionIDs["create_roles"]},
		{ids.ClientID, ids.PermissionIDs["update_roles"]},
		{ids.ClientID, ids.PermissionIDs["delete_roles"]},
		{ids.ClientID, ids.PermissionIDs["read_clients"]},
		{ids.ClientID, ids.PermissionIDs["create_clients"]},
		{ids.ClientID, ids.PermissionIDs["update_clients"]},
		{ids.ClientID, ids.PermissionIDs["delete_clients"]},
	}
	for _, cp := range clientPermissions {
		_, err := tx.ExecContext(ctx, `INSERT INTO client_permissions (client_id, permission_id) VALUES (?, ?)`, cp.ClientID, cp.PermissionID)
		if err != nil {
			return err
		}
	}
	return nil
}
