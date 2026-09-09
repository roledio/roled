-- +goose Up
ALTER TABLE user_identities ADD COLUMN project_id CHAR(22) NULL; -- Set nullable initially

-- Update missing project_id in table user_identities
UPDATE user_identities 
SET project_id = (SELECT project_id
    FROM users
    WHERE users.id = user_identities.user_id)
WHERE project_id IS NULL;

-- Set project_id column to NOT NULL
ALTER TABLE user_identities
MODIFY COLUMN project_id CHAR(22) NOT NULL;

-- Temporarily disable foreign key check for DROP INDEX
-- to avoid error when dropping index that is part of a foreign key
SET FOREIGN_KEY_CHECKS = 0;

-- Drop old unique key by provider and provider_user_id
ALTER TABLE user_identities 
DROP INDEX uk_user_identities_provider_provider_user_id,
DROP INDEX idx_user_identities_user_id_deleted_at;

-- Add unique key by provider, provider_user_id, and project_id
CREATE UNIQUE INDEX uk_provider_provider_user_id_project_id 
ON user_identities(provider, provider_user_id, project_id);

CREATE INDEX idx_project_id_user_id_deleted_at ON user_identities(project_id, user_id, deleted_at);

-- Re-enable foreign key check
SET FOREIGN_KEY_CHECKS = 1;

-- +goose Down
ALTER TABLE user_identities DROP INDEX idx_project_id_user_id_deleted_at;

ALTER TABLE user_identities DROP INDEX uk_provider_provider_user_id_project_id;

CREATE UNIQUE INDEX uk_user_identities_provider_provider_user_id 
ON user_identities(provider, provider_user_id);

CREATE INDEX idx_user_identities_user_id_deleted_at 
ON user_identities(user_id, deleted_at);

ALTER TABLE user_identities DROP COLUMN project_id;
