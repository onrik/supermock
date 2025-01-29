package handlers

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/onrik/supermock/pkg/models"

	"github.com/labstack/echo/v4"
)

type DB interface {
	Requests(ctx context.Context, testID string) ([]models.Request, error)
	Response(ctx context.Context, method, path string) (*models.Response, error)
	Responses(ctx context.Context) ([]models.Response, error)
	ResponseDelete(ctx context.Context, uuid string) error
	ResponseSave(ctx context.Context, response models.Response) error
	SaveRequest(ctx context.Context, request models.Request) error
	Clean(ctx context.Context, testID string) error
}

type SMTP interface {
	Emails() []models.Email
	Purge()
}

type Handlers struct {
	db   DB
	smtp SMTP
}

func New(db DB, smtp SMTP) *Handlers {
	return &Handlers{
		db:   db,
		smtp: smtp,
	}
}

/*
Requests
@openapi GET /_requests/{test_id}
@openapiParam test_id in=path, type=string, example=194a0bde-d70f-4b16-a303-1ffa2a77c143
@openapiSummary Get requests for test
@openapiResponse 200 application/json {"requests": []models.Request}
*/
func (h *Handlers) Requests(c echo.Context) error {
	testID := c.Param("test_id")
	requests, err := h.db.Requests(c.Request().Context(), testID)
	if err != nil {
		slog.Error("Get requests error", "error", err, "test_id", testID)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, echo.Map{
		"requests": requests,
	})
}

/*
ResponseCreate save response
@openapi POST /_responses
@openapiSummary Put response
@openapiRequest application/json models.Response
@openapiResponse 400 application/json {"message": "uuid=required,test_id=required,method=required,path=required,status=required"}
@openapiResponse 200 application/json {}
*/
func (h *Handlers) ResponseCreate(c echo.Context) error {
	response := models.Response{}
	err := c.Bind(&response)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	err = c.Validate(&response)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	err = h.db.ResponseSave(c.Request().Context(), response)
	if err != nil {
		slog.Error("Save response error", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	slog.Debug("Response saved",
		"id", response.ID,
		"uuid", response.UUID,
		"test_id", response.TestID,
		"method", response.Method,
		"path", response.Path,
		"status", response.Status,
		"headers", response.Headers,
	)

	return c.JSON(http.StatusOK, echo.Map{})
}

/*
ResponsesList
@openapi GET /_responses
@openapiSummary Get responses
@openapiResponse 200 application/json {"responses": []models.Response}
*/
func (h *Handlers) ResponseList(c echo.Context) error {
	responses, err := h.db.Responses(c.Request().Context())
	if err != nil {
		slog.Error("Get responses error", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, echo.Map{
		"responses": responses,
	})
}

/*
DeleteResponse
@openapi DELETE /_responses/{uuid}
@openapiParam uuid in=path, type=string, example=2b78ffe3-ce2b-46e3-ae71-6509e1613068
@openapiResponse 200 application/json {}
*/
func (h *Handlers) DeleteResponse(c echo.Context) error {
	uuid := c.Param("uuid")
	err := h.db.ResponseDelete(c.Request().Context(), uuid)
	if err != nil {
		slog.Error("Delete response error", "error", err, "uuid", uuid)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	slog.Debug("Response deleted", "uuid", uuid)

	return c.JSON(http.StatusOK, echo.Map{})
}

/*
Clean
@openapi DELETE /_tests/{test_id}
@openapiSummary Delete all requests and responses by test id
@openapiParam test_id in=path, type=string, example=d3fb230c-c9d9-4e7a-b936-15b6c6c891aa
@openapiResponse 200 application/json {}
*/
func (h *Handlers) Clean(c echo.Context) error {
	testID := c.Param("test_id")
	err := h.db.Clean(c.Request().Context(), testID)
	if err != nil {
		slog.Error("Clean test error", "test_id", testID, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	slog.Info("Test cleaned", "test_id", testID)

	return c.JSON(http.StatusOK, echo.Map{})
}

func (h *Handlers) Catch(c echo.Context) error {
	method := c.Request().Method
	path := c.Request().URL.Path

	slog.Debug(fmt.Sprintf("Request<- %s %s", method, path))

	response, err := h.db.Response(c.Request().Context(), method, path)
	if err != nil {
		slog.Error("Get response error", "error", err, "method", method, "path", path)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	defer c.Request().Body.Close()

	if response == nil {
		return c.NoContent(http.StatusNotImplemented)
	}

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	if !response.DisableCatch {
		request := models.Request{
			TestID:  response.TestID,
			Method:  c.Request().Method,
			Path:    c.Request().URL.Path,
			Body:    string(body),
			Headers: map[string]string{},
		}

		for k := range c.Request().Header {
			request.Headers[k] = c.Request().Header.Get(k)
		}

		err = h.db.SaveRequest(c.Request().Context(), request)
		if err != nil {
			slog.Error("Save request error", "error", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}

		slog.Info("Request saved", "method", method, "path", path, "test_id", request.TestID)
	}

	c.Response().Status = int(response.Status)
	for k, v := range response.Headers {
		c.Response().Header().Set(k, v)
	}
	_, err = c.Response().Write([]byte(response.Body))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	slog.Debug(fmt.Sprintf("Response-> %s %s", method, path), "status", response.Status)

	return nil
}
