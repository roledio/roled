-- +goose Up
CREATE TABLE accounts (
    id CHAR(22) NOT NULL,
    created_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4),
    updated_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4) ON UPDATE CURRENT_TIMESTAMP(4),
    deleted_at TIMESTAMP(4) NULL,
    name VARCHAR(128) NOT NULL,
    description VARCHAR(512),
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (id),
    KEY idx_accounts_is_active_deleted_at_updated_at (is_active, deleted_at, updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- members: users who manage Roled account (account's members)
CREATE TABLE members (
    id CHAR(22) NOT NULL,
    created_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4),
    updated_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4) ON UPDATE CURRENT_TIMESTAMP(4),
    deleted_at TIMESTAMP(4) NULL,
    account_id CHAR(22) NOT NULL,
    user_id CHAR(22) NOT NULL,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE, -- indicates if the member is an admin of the account
    PRIMARY KEY (id),
    KEY idx_members_account_id_deleted_at_user_id_updated_at (account_id, deleted_at, user_id, updated_at),
    CONSTRAINT fk_members_account_id FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE projects (
    id CHAR(22) NOT NULL,
    created_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4),
    updated_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4) ON UPDATE CURRENT_TIMESTAMP(4),
    deleted_at TIMESTAMP(4) NULL,
    account_id CHAR(22) NOT NULL,
    name VARCHAR(128) NOT NULL,
    description VARCHAR(512),
    logo_url VARCHAR(512),
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (id),
    KEY idx_projects_account_id_system_deleted_at_updated_at (account_id, is_system, deleted_at, updated_at),
    KEY idx_projects_is_system_deleted_at (is_system, deleted_at),
    CONSTRAINT fk_projects_account_id FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE redirect_uris (
    created_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4),
    project_id CHAR(22) NOT NULL,
    redirect_uri VARCHAR(512) NOT NULL,
    login_url VARCHAR(512) NULL,
    PRIMARY KEY (project_id, redirect_uri),
    CONSTRAINT fk_redirect_uris_project_id FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE project_settings (
    id CHAR(22) NOT NULL,
    created_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4),
    updated_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4) ON UPDATE CURRENT_TIMESTAMP(4),
    project_id CHAR(22) NOT NULL,
    is_signup_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    default_signup_role_id CHAR(22),
    is_signup_verify_email BOOLEAN NOT NULL DEFAULT FALSE,
    is_forgot_password_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    is_allow_temp_email BOOLEAN NOT NULL DEFAULT TRUE, -- allow temporary/disposable email addresses
    PRIMARY KEY (id),
    KEY idx_project_settings_project_id (project_id),
    CONSTRAINT fk_project_settings_project_id FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE resources (
    id CHAR(22) NOT NULL,
    created_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4),
    updated_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4) ON UPDATE CURRENT_TIMESTAMP(4),
    account_id CHAR(22) NOT NULL,
    project_id CHAR(22) NOT NULL,
    code VARCHAR(64) NOT NULL, -- must be unique per project (checked in application logic)
    name VARCHAR(128) NOT NULL,
    description VARCHAR(512),
    is_default BOOLEAN NOT NULL DEFAULT FALSE, -- whether this resource is included by default in new projects
    PRIMARY KEY (id),
    UNIQUE KEY uk_resources_project_id_code (project_id, code),
    KEY idx_resources_project_id (project_id),
    CONSTRAINT fk_resources_account_id FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
    CONSTRAINT fk_resources_project_id FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE permissions (
    id CHAR(22) NOT NULL,
    created_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4),
    updated_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4) ON UPDATE CURRENT_TIMESTAMP(4),
    resource_id CHAR(22) NOT NULL,
    code VARCHAR(64) NOT NULL, -- must be unique per resource (checked in application logic)
    name VARCHAR(128) NOT NULL,
    description VARCHAR(512),
    is_default BOOLEAN NOT NULL DEFAULT FALSE, -- whether this permission is included by default in new projects
    PRIMARY KEY (id),
    UNIQUE KEY uk_permissions_resource_id_code (resource_id, code),
    KEY idx_permissions_resource_id (resource_id),
    CONSTRAINT fk_permissions_resource_id FOREIGN KEY (resource_id) REFERENCES resources(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE roles (
    id CHAR(22) NOT NULL,
    created_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4),
    updated_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4) ON UPDATE CURRENT_TIMESTAMP(4),
    deleted_at TIMESTAMP(4) NULL,
    account_id CHAR(22) NOT NULL,
    project_id CHAR(22) NOT NULL,
    code VARCHAR(64) NOT NULL, -- must be unique per project (checked in application logic)
    name VARCHAR(128) NOT NULL,
    description VARCHAR(512),
    PRIMARY KEY (id),
    KEY idx_roles_project_id_deleted_at_code (project_id, deleted_at, code),
    CONSTRAINT fk_roles_account_id FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
    CONSTRAINT fk_roles_project_id FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Add foreign key for default_signup_role_id in project_settings table after roles table is created
ALTER TABLE project_settings
ADD CONSTRAINT fk_project_settings_default_signup_role_id 
FOREIGN KEY (default_signup_role_id) 
REFERENCES roles(id);

CREATE TABLE role_permissions (
    created_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4),
    role_id CHAR(22) NOT NULL,
    permission_id CHAR(22) NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    CONSTRAINT fk_role_permissions_role_id FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    CONSTRAINT fk_role_permissions_permission_id FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- clients: clients belonging to a project for OAuth2 client_credentials grant
CREATE TABLE clients (
    id CHAR(22) NOT NULL,
    created_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4),
    updated_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4) ON UPDATE CURRENT_TIMESTAMP(4),
    deleted_at TIMESTAMP(4) NULL,
    account_id CHAR(22) NOT NULL,
    project_id CHAR(22) NOT NULL,
    name VARCHAR(128) NOT NULL,
    description VARCHAR(512),
    secret_encrypted VARCHAR(512) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (id),
    KEY idx_clients_project_id_is_default_deleted_at (project_id, is_default, deleted_at),
    KEY idx_clients_project_id_deleted_at_is_active_updated_at (project_id, deleted_at, is_active, updated_at),
    CONSTRAINT fk_clients_account_id FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
    CONSTRAINT fk_clients_project_id FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE client_permissions (
    created_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4),
    client_id CHAR(22) NOT NULL,
    permission_id CHAR(22) NOT NULL,
    PRIMARY KEY (client_id, permission_id),
    KEY idx_client_permissions_client_id (client_id),
    CONSTRAINT fk_client_permissions_client_id FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE,
    CONSTRAINT fk_client_permissions_permission_id FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- users: end-users belonging to a project == users of the client apps
CREATE TABLE users (
    id CHAR(22) NOT NULL,
    created_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4),
    updated_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4) ON UPDATE CURRENT_TIMESTAMP(4),
    deleted_at TIMESTAMP(4) NULL,
    account_id CHAR(22) NOT NULL,
    project_id CHAR(22) NOT NULL,
    email VARCHAR(128), -- unique per project (checked in application logic)
    email_verified_at TIMESTAMP(4) NULL,
    password_hash VARCHAR(256), -- might be null if using external identity provider
    external_user_id VARCHAR(128), -- user id from external identity provider, either email or this must be provided
    display_name VARCHAR(128) NOT NULL,
    avatar_url VARCHAR(512),
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (id),
    KEY idx_users_project_id_deleted_at_created_at (project_id, deleted_at, created_at),
    KEY idx_users_project_id_email_deleted_at (project_id, email, deleted_at),
    KEY idx_users_project_id_external_user_id_deleted_at (project_id, external_user_id, deleted_at),
    CONSTRAINT fk_users_account_id FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
    CONSTRAINT fk_users_project_id FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Add foreign key for user_id in members table after users table is created
ALTER TABLE members
ADD CONSTRAINT fk_members_user_id 
FOREIGN KEY (user_id) 
REFERENCES users(id);

-- user_roles: many-to-many relationship between users and roles
CREATE TABLE user_roles (
    created_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4),
    user_id CHAR(22) NOT NULL,
    role_id CHAR(22) NOT NULL,
    PRIMARY KEY (user_id, role_id),
    KEY idx_user_roles_user_id (user_id),
    KEY idx_user_roles_role_id (role_id),
    CONSTRAINT fk_user_roles_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_roles_role_id FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE refresh_tokens (
    id CHAR(22) NOT NULL,
    created_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4),
    updated_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4) ON UPDATE CURRENT_TIMESTAMP(4),
    account_id CHAR(22) NOT NULL,
    project_id CHAR(22) NOT NULL,
    client_id CHAR(22) NOT NULL,
    user_id CHAR(22),
    refresh_token_hash VARCHAR(256) NOT NULL,
    status ENUM('issued', 'used', 'expired', 'revoked') NOT NULL,
    expires_in INT, -- in seconds e.g. 3600 for 1 hour
    used_at TIMESTAMP(4) NULL,
    issued_at TIMESTAMP(4) NULL,
    revoked_at TIMESTAMP(4) NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_refresh_tokens_client_id_refresh_token_hash (client_id, refresh_token_hash),
    CONSTRAINT fk_refresh_tokens_account_id FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
    CONSTRAINT fk_refresh_tokens_project_id FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    CONSTRAINT fk_refresh_tokens_client_id FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE,
    CONSTRAINT fk_refresh_tokens_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE auth_codes (
    id CHAR(22) NOT NULL,
    created_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4),
    updated_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4) ON UPDATE CURRENT_TIMESTAMP(4),
    account_id CHAR(22) NOT NULL,
    project_id CHAR(22) NOT NULL,
    client_id CHAR(22) NOT NULL,
    user_id CHAR(22),
    code_hash VARCHAR(256) NOT NULL,
    code_challenge VARCHAR(512) NOT NULL, -- hashed using SHA256 and base64rawurl-encoded
    code_challenge_method ENUM('S256') NOT NULL, -- only support S256
    redirect_uri VARCHAR(256) NOT NULL, -- to verify with the one used in /authorize page
    state VARCHAR(512),
    expires_at TIMESTAMP(4) NOT NULL,
    used_at TIMESTAMP(4) NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_auth_codes_client_id_code_hash (client_id, code_hash),
    CONSTRAINT fk_auth_codes_account_id FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
    CONSTRAINT fk_auth_codes_project_id FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    CONSTRAINT fk_auth_codes_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_auth_codes_client_id FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE access_tokens (
    id CHAR(22) NOT NULL,
    created_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4),
    updated_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4) ON UPDATE CURRENT_TIMESTAMP(4),
    deleted_at TIMESTAMP(4) NULL,
    account_id CHAR(22) NOT NULL,
    project_id CHAR(22) NOT NULL,
    client_id CHAR(22) NOT NULL,
    user_id CHAR(22),
    refresh_token_id CHAR(22),
    auth_code_id CHAR(22),
    grant_type ENUM('client_credentials', 'authorization_code', 'refresh_token') NOT NULL,
    status ENUM('issued', 'expired', 'revoked') NOT NULL,
    expires_in INT, -- in seconds e.g., 3600 for 1 hour
    issued_at TIMESTAMP(4) NULL,
    revoked_at TIMESTAMP(4) NULL,
    PRIMARY KEY (id),
    KEY idx_access_tokens_account_id_deleted_at (account_id, deleted_at),
    KEY idx_access_tokens_project_id_deleted_at (project_id, deleted_at),
    KEY idx_access_tokens_client_id_deleted_at (client_id, deleted_at),
    KEY idx_access_tokens_user_id_deleted_at (user_id, deleted_at),
    CONSTRAINT fk_access_tokens_account_id FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
    CONSTRAINT fk_access_tokens_project_id FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    CONSTRAINT fk_access_tokens_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_access_tokens_client_id FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE,
    CONSTRAINT fk_access_tokens_refresh_token_id FOREIGN KEY (refresh_token_id) REFERENCES refresh_tokens(id) ON DELETE SET NULL,
    CONSTRAINT fk_access_tokens_auth_code_id FOREIGN KEY (auth_code_id) REFERENCES auth_codes(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


-- +goose Down
-- We should never do rollback migration in production which is risky and can lead to data loss.
-- It is generally considered best practice to avoid automatic or automatic-down database migration rollbacks in production. 
-- Instead, the industry standard is to "roll forward" by applying a new, corrective migration to fix the issue.
-- This down migration is provided for testing and local development purposes only, and should not be used in production environments.
DROP TABLE IF EXISTS access_tokens;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS auth_codes;
DROP TABLE IF EXISTS client_permissions;
DROP TABLE IF EXISTS clients;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS project_settings;
DROP TABLE IF EXISTS redirect_uris;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS members;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS resources;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS accounts;
