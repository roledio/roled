-- +goose Up
CREATE TABLE brandings (
    project_id CHAR(22) NOT NULL,
    created_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4),
    updated_at TIMESTAMP(4) NOT NULL DEFAULT CURRENT_TIMESTAMP(4) ON UPDATE CURRENT_TIMESTAMP(4),
    logo_url VARCHAR(2048) NULL,
    favicon_url VARCHAR(2048) NULL,
    primary_color CHAR(7) NOT NULL DEFAULT '#ba8d1c',
    rounding ENUM('sharp', 'small', 'medium', 'large') NOT NULL DEFAULT 'small',
    enable_shadow BOOLEAN NOT NULL DEFAULT TRUE,
    enable_border BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (project_id),
    CONSTRAINT fk_project_id FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO brandings (project_id, logo_url)
SELECT id, logo_url FROM projects WHERE is_system = TRUE AND deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS brandings;
