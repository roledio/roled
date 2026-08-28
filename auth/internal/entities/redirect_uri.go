package entities

import (
	"time"
)

type RedirectURI struct {
	CreatedAt   time.Time `db:"created_at"`
	ProjectID   string    `db:"project_id"`
	RedirectURI string    `db:"redirect_uri"`
	LoginURL    *string   `db:"login_url"`
}
