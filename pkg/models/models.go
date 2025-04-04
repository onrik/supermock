package models

type Request struct {
	TestID    string            `json:"test_id" openapi:"format=uuid"`
	Method    string            `json:"method"`
	Query     string            `json:"query"`
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
	Status       uint16            `json:"status" validate:"required"`
	Headers      map[string]string `json:"headers"`
	Body         string            `json:"body"`
	IsPermanent  bool              `json:"is_permanent"`
	DisableCatch bool              `json:"disable_catch"`
}

type Email struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Date        string `json:"date"`
	Subject     string `json:"subject"`
	ContentType string `json:"content_type"`
	Body        string `json:"body"`
	Raw         string `json:"raw"`
}
