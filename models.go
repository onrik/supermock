package main

type Request struct {
	TestID    string            `json:"test_id" openapi:"format=uuid"`
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Headers   map[string]string `json:"headers"`
	Body      string            `json:"body"`
	CreatedAt string            `json:"created_at" openapi:"format=date-time"`
}

type Response struct {
	ID           int64             `json:"-"`
	UUID         string            `json:"uuid" validate:"required" openapi:"format=uuid"`
	TestID       string            `json:"test_id" validate:"required" openapi:"format=uuid"`
	Method       string            `json:"method" validate:"required"`
	Path         string            `json:"path" validate:"required"`
	Status       uint              `json:"status" validate:"required"`
	Headers      map[string]string `json:"headers"`
	Body         string            `json:"body"`
	IsPermanent  bool              `json:"is_permanent"`
	DisableCatch bool              `json:"disable_catch"`
}
