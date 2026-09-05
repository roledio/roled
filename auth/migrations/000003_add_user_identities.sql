-- +goose Up
CREATE TABLE user_identities (
    id CHAR(22) NOT NULL,
    created_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4),
    updated_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4) ON UPDATE CURRENT_TIMESTAMP(4),
    deleted_at TIMESTAMP(4) NULL,
    user_id CHAR(22) NOT NULL,
    provider VARCHAR(64) NOT NULL,
    provider_user_id VARCHAR(256) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_user_identities_provider_provider_user_id (provider, provider_user_id),
    KEY idx_user_identities_user_id_deleted_at (user_id, deleted_at),
    CONSTRAINT fk_user_identities_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS user_identities;
