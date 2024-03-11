package db

import (
	"database/sql"
	"fmt"
	"net/url"
)

func initSqlite(dsn url.URL) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?cache=shared", dsn.Host))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS requests (
		id INTEGER NOT NULL PRIMARY KEY,
		test_id TEXT NOT NULL,
		method TEXT NOT NULL,
		path TEXT NOT NULL,
		headers TEXT NOT NULL,
		body TEXT NOT NULL,
		created_at TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS responses (
		id INTEGER NOT NULL PRIMARY KEY,
		uuid TEXT NOT NULL,
		test_id TEXT NOT NULL,
		method TEXT NOT NULL,
		path TEXT NOT NULL,
		status INTEGER NOT NULL,
		headers TEXT NOT NULL,
		body TEXT NOT NULL,
		is_permanent bool NOT NULL,
		disable_catch bool NOT NULL,
		created_at TEXT NOT NULL
	);`)

	return db, err
}
