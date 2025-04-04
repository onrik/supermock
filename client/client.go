package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Request struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Query   string            `json:"query"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

type Response struct {
	UUID         string            `json:"uuid"`
	TestID       string            `json:"test_id"`
	Method       string            `json:"method"`
	Path         string            `json:"path"`
	Status       uint              `json:"status"`
	Headers      map[string]string `json:"headers"`
	Body         string            `json:"body"`
	IsPermanent  bool              `json:"is_permanent"`
	DisableCatch bool              `json:"disable_catch"`
}

type Client struct {
	url  string
	http *http.Client
}

func New(baseURL string, client *http.Client) *Client {
	if client == nil {
		client = &http.Client{
			Timeout: 5 * time.Second,
		}
	}
	return &Client{
		url:  strings.TrimRight(baseURL, "/"),
		http: client,
	}
}

func (c *Client) putReponse(ctx context.Context, r Response) error {
	body, err := json.Marshal(r)
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url+"/_responses", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	request.Header.Add("Content-Type", "application/json")

	response, err := c.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode >= 400 {
		return fmt.Errorf("http status: %d", response.StatusCode)
	}

	return nil
}

// Put response to stack
func (c *Client) Put(ctx context.Context, responses ...Response) error {
	for i := range responses {
		err := c.putReponse(ctx, responses[i])
		if err != nil {
			return err
		}
	}

	return nil
}

// Get requests by test id
func (c *Client) Get(ctx context.Context, testID string) ([]Request, error) {
	url := fmt.Sprintf("%s/_requests/%s", c.url, testID)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "application/json")

	response, err := c.http.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode >= 400 {
		return nil, fmt.Errorf("http status: %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	r := struct {
		Requests []Request `json:"requests"`
	}{}

	err = json.Unmarshal(body, &r)

	return r.Requests, err
}

// Clean test requests and responses
func (c *Client) Clean(ctx context.Context, testID string) error {
	url := fmt.Sprintf("%s/_tests/%s", c.url, testID)
	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	request.Header.Add("Content-Type", "application/json")

	response, err := c.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode >= 400 {
		return fmt.Errorf("http status: %d", response.StatusCode)
	}

	return nil
}
