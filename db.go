package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"time"
)

type DB struct {
	sql *sql.DB
}

func NewDB(dsn string) (*DB, error) {
	parsedDSN, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}

	var db *sql.DB

	if parsedDSN.Scheme == "sqlite3" || parsedDSN.Scheme == "sqlite" {
		db, err = sql.Open("sqlite3", fmt.Sprintf("file:%s?cache=shared", parsedDSN.Host))
		if err != nil {
			return nil, err
		}
		db.SetMaxOpenConns(1)
	} else if parsedDSN.Scheme == "postgres" {
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			return nil, err
		}
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

	return &DB{db}, err
}

func (db *DB) Close() {
	err := db.sql.Close()
	if err != nil {
		slog.Error("Close db error", "error", err)
	}
}

func (db *DB) Requests(ctx context.Context, testID string) ([]Request, error) {
	requests := []Request{}
	rows, err := db.sql.QueryContext(ctx, "SELECT test_id, method, path, headers, body, created_at FROM requests WHERE test_id = $1", testID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		request := Request{
			Headers: map[string]string{},
		}
		var headers string
		err = rows.Scan(&request.TestID, &request.Method, &request.Path, &headers, &request.Body, &request.CreatedAt)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}

	return requests, nil
}

func (db *DB) Response(ctx context.Context, method, path string) (*Response, error) {
	rows, err := db.sql.QueryContext(
		ctx,
		"SELECT id, test_id, status, headers, body, is_permanent, disable_catch FROM responses WHERE method = $1 AND path = $2 ORDER BY id ASC LIMIT 1",
		method, path)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}

	if !rows.Next() {
		rows.Close()
		return nil, nil
	}

	response := Response{
		Headers: map[string]string{},
	}
	var headers string
	err = rows.Scan(&response.ID, &response.TestID, &response.Status, &headers, &response.Body, &response.IsPermanent, &response.DisableCatch)
	if err != nil {
		return nil, err
	}

	if len(headers) > 0 {
		err = json.Unmarshal([]byte(headers), &response.Headers)
		if err != nil {
			return nil, err
		}
	}

	fmt.Printf("%+v\n", response)
	rows.Close()

	if !response.IsPermanent {
		_, err = db.sql.ExecContext(ctx, "DELETE FROM responses WHERE id = $1", response.ID)
		if err != nil {
			return nil, fmt.Errorf("delete error: %w", err)
		}
	}

	return &response, nil
}

func (db *DB) DeleteResponse(ctx context.Context, uuid string) error {
	_, err := db.sql.ExecContext(ctx, "DELETE FROM responses WHERE uuid = $1", uuid)
	return err
}

func (db *DB) SaveResponse(ctx context.Context, response Response) error {
	if response.Headers == nil {
		response.Headers = map[string]string{}
	}
	headers, err := json.Marshal(response.Headers)
	if err != nil {
		return err
	}

	_, err = db.sql.Exec(
		"INSERT INTO responses (uuid, test_id, method, path, status, headers, body, is_permanent, disable_catch, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
		response.UUID, response.TestID, response.Method, response.Path, response.Status, string(headers), response.Body, response.IsPermanent, response.DisableCatch, time.Now().UTC().Format(time.RFC3339))
	return err
}

func (db *DB) SaveRequest(ctx context.Context, request Request) error {
	headers, err := json.Marshal(request.Headers)
	if err != nil {
		return err
	}
	request.CreatedAt = time.Now().UTC().Format(time.RFC3339)

	_, err = db.sql.Exec(
		"INSERT INTO requests (test_id, method, path, headers, body, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		request.TestID, request.Method, request.Path, string(headers), request.Body, request.CreatedAt)

	return err
}

func (db *DB) Clean(ctx context.Context, testID string) error {
	_, err := db.sql.ExecContext(ctx, "DELETE FROM requests WHERE test_id = $1", testID)
	if err != nil {
		return err
	}

	_, err = db.sql.ExecContext(ctx, "DELETE FROM responses WHERE test_id = $1", testID)
	if err != nil {
		return err
	}

	return nil
}
