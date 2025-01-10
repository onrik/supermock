package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/onrik/supermock/pkg/models"
)

type DB struct {
	sql *sql.DB
}

func New(dsn string) (*DB, error) {
	parsedDSN, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}

	if parsedDSN.Scheme == "sqlite3" || parsedDSN.Scheme == "sqlite" {
		db, err := initSqlite(*parsedDSN)
		return &DB{db}, err
	} else if parsedDSN.Scheme == "postgres" {
		db, err := initPostgresql(*parsedDSN)
		return &DB{db}, err
	}
	return nil, fmt.Errorf("unsupported dsn scheme: %s", parsedDSN.Scheme)
}

func (db *DB) Close() {
	err := db.sql.Close()
	if err != nil {
		slog.Error("Close db error", "error", err)
	}
}

func (db *DB) Requests(ctx context.Context, testID string) ([]models.Request, error) {
	requests := []models.Request{}
	rows, err := db.sql.QueryContext(ctx, "SELECT test_id, method, path, headers, body, created_at FROM requests WHERE test_id = $1", testID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		request := models.Request{
			Headers: map[string]string{},
		}
		var headers string
		err = rows.Scan(&request.TestID, &request.Method, &request.Path, &headers, &request.Body, &request.CreatedAt)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(headers), &request.Headers)
		if err != nil {
			return nil, err
		}

		requests = append(requests, request)
	}

	return requests, nil
}

func (db *DB) Response(ctx context.Context, method, path string) (*models.Response, error) {
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

	response := models.Response{
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

func (db *DB) SaveResponse(ctx context.Context, response models.Response) error {
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

func (db *DB) SaveResponses(ctx context.Context, responses ...models.Response) error {
	for i := range responses {
		err := db.SaveResponse(ctx, responses[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) SaveRequest(ctx context.Context, request models.Request) error {
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
