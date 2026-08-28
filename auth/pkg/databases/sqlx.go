package databases

import "github.com/jmoiron/sqlx"

func NewSqlx(config Config) (*sqlx.DB, error) {
	db, err := Open(config)
	if err != nil {
		return nil, err
	}
	return sqlx.NewDb(db, config.Driver), nil
}
