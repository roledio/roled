-- +goose Up
ALTER TABLE oauth_connections 
ADD COLUMN credential_type ENUM('default', 'custom') NOT NULL,
MODIFY COLUMN client_id VARCHAR(512) NULL,          -- nullable for default credentials
MODIFY COLUMN client_secret_encrypted TEXT NULL,    -- nullable for default credentials
MODIFY COLUMN scopes TEXT NULL;                     -- nullable for default credentials

-- +goose Down
ALTER TABLE oauth_connections 
DROP COLUMN credential_type,
MODIFY COLUMN client_id VARCHAR(512) NOT NULL,
MODIFY COLUMN client_secret_encrypted VARCHAR(512) NOT NULL,
MODIFY COLUMN scopes VARCHAR(512) NOT NULL;
