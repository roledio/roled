-- +goose Up
UPDATE brandings b
JOIN projects p ON p.id = b.project_id
SET b.favicon_url = REPLACE(p.logo_url, 'roled-logo.png', 'favicon.ico')
WHERE p.is_system = TRUE
  AND p.deleted_at IS NULL
  AND p.logo_url IS NOT NULL;

-- +goose Down
UPDATE brandings b
JOIN projects p ON p.id = b.project_id
SET b.favicon_url = NULL
WHERE p.is_system = TRUE
  AND p.deleted_at IS NULL
  AND b.favicon_url = REPLACE(p.logo_url, 'roled-logo.png', 'favicon.ico');
