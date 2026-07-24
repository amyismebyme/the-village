package database

import "database/sql"

type Database struct {
	DB *sql.DB
}

func New(db *sql.DB) *Database {
	return &Database{
		DB: db,
	}
}
