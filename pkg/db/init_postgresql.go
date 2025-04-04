package db

import (
	"database/sql"
	"net/url"
)

func initPostgresql(dsn url.URL) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn.String())
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS requests (
		id SERIAL PRIMARY KEY,
		test_id TEXT NOT NULL,
		method TEXT NOT NULL,
		path TEXT NOT NULL,
		headers TEXT NOT NULL,
		body TEXT NOT NULL,
		created_at TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS responses (
		id SERIAL PRIMARY KEY,
		uuid TEXT NOT NULL UNIQUE,
		test_id TEXT NOT NULL,
		method TEXT NOT NULL,
		path TEXT NOT NULL,
		status INTEGER NOT NULL,
		headers TEXT NOT NULL,
		body TEXT NOT NULL,
		is_permanent bool NOT NULL,
		disable_catch bool NOT NULL,
		created_at TEXT NOT NULL
	);

	ALTER TABLE requests ADD COLUMN IF NOT EXISTS query TEXT NOT NULL DEFAULT '';
`)

	return db, err
}
