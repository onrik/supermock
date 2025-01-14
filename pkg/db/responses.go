package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/onrik/supermock/pkg/models"
)

func (db *DB) Response(ctx context.Context, method, path string) (*models.Response, error) {
	rows, err := db.sql.QueryContext(
		ctx,
		"SELECT id, test_id, status, headers, body, is_permanent, disable_catch FROM responses WHERE method = $1 AND path = $2 ORDER BY id ASC LIMIT 1",
		method, path)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}

	defer rows.Close()

	if !rows.Next() {
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

	rows.Close()

	if !response.IsPermanent {
		_, err = db.sql.ExecContext(ctx, "DELETE FROM responses WHERE id = $1", response.ID)
		if err != nil {
			return nil, fmt.Errorf("delete error: %w", err)
		}

		slog.InfoContext(ctx, "Response deleted", "id", response.ID, "test_id", response.TestID)
	}

	return &response, nil
}

func (db *DB) Responses(ctx context.Context) ([]models.Response, error) {
	rows, err := db.sql.QueryContext(
		ctx,
		"SELECT id, test_id, method, path, status, headers, body, is_permanent, disable_catch FROM responses",
	)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}

	defer rows.Close()

	responses := []models.Response{}
	for rows.Next() {
		response := models.Response{
			Headers: map[string]string{},
		}
		var headers string
		err = rows.Scan(
			&response.ID,
			&response.TestID,
			&response.Method,
			&response.Path,
			&response.Status,
			&headers,
			&response.Body,
			&response.IsPermanent,
			&response.DisableCatch,
		)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		if len(headers) > 0 {
			err = json.Unmarshal([]byte(headers), &response.Headers)
			if err != nil {
				slog.ErrorContext(ctx, "Unmarshal response headers error", "error", err, "response", response, "headers", headers)
			}
		}

		responses = append(responses, response)
	}

	return responses, nil
}

func (db *DB) ResponseDelete(ctx context.Context, uuid string) error {
	_, err := db.sql.ExecContext(ctx, "DELETE FROM responses WHERE uuid = $1", uuid)
	return err
}

func (db *DB) ResponseSave(ctx context.Context, response models.Response) error {
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

func (db *DB) ResponsesSave(ctx context.Context, responses ...models.Response) error {
	for i := range responses {
		err := db.ResponseSave(ctx, responses[i])
		if err != nil {
			return err
		}
	}

	return nil
}
