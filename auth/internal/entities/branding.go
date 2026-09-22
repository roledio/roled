package entities

import "time"

type Branding struct {
	ProjectID    string    `db:"project_id"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	LogoURL      *string   `db:"logo_url"`
	FaviconURL   *string   `db:"favicon_url"`
	PrimaryColor string    `db:"primary_color"`
	Rounding     string    `db:"rounding"`
	EnableShadow bool      `db:"enable_shadow"`
	EnableBorder bool      `db:"enable_border"`
}
