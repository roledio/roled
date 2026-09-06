-- +goose Up
CREATE TABLE oauth_connections (
    id CHAR(22) NOT NULL,
    created_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4),
    updated_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4) ON UPDATE CURRENT_TIMESTAMP(4),
    project_id CHAR(22) NOT NULL,
    provider VARCHAR(32) NOT NULL,
    client_id VARCHAR(512) NOT NULL,
    client_secret_encrypted VARCHAR(512) NOT NULL,
    scopes VARCHAR(512) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    PRIMARY KEY (id),
    KEY idx_oauth_connections_project_id_provider (project_id, provider),
    CONSTRAINT fk_oauth_connections_project_id FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS oauth_connections;
