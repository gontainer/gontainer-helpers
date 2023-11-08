package repositories

import (
	"database/sql"
)

// UserRepository is a sample struct that may perform queries on the DB
type UserRepository struct {
	Tx *sql.Tx
}

// ImageRepository is a sample struct that may perform queries on the DB
type ImageRepository struct {
	Tx *sql.Tx
}
