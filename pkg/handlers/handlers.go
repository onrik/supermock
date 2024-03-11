package handlers

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/onrik/supermock/pkg/models"

	"github.com/labstack/echo/v4"
)

type DB interface {
	Requests(ctx context.Context, testID string) ([]models.Request, error)
	Response(ctx context.Context, method, path string) (*models.Response, error)
	DeleteResponse(ctx context.Context, uuid string) error
	SaveResponse(ctx context.Context, response models.Response) error
	SaveRequest(ctx context.Context, request models.Request) error
	Clean(ctx context.Context, testID string) error
}

type Handlers struct {
	db DB
}

func New(db DB) *Handlers {
	return &Handlers{
		db: db,
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
Responses save response
@openapi POST /_responses
@openapiSummary Put response
@openapiRequest application/json models.Response
@openapiResponse 400 application/json {"message": "uuid=required,test_id=required,method=required,path=required,status=required"}
@openapiResponse 200 application/json {}
*/
func (h *Handlers) Responses(c echo.Context) error {
	response := models.Response{}
	err := c.Bind(&response)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	err = c.Validate(&response)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	slog.Debug("Save response",
		"UUID", response.UUID,
		"test_id", response.TestID,
		"method", response.Method,
		"path", response.Path,
		"status", response.Status,
	)
	err = h.db.SaveResponse(c.Request().Context(), response)
	if err != nil {
		slog.Error("Save response error", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, echo.Map{})
}

/*
DeleteResponse
@openapi DELETE /_responses/{uuid}
@openapiParam uuid in=path, type=string, example=2b78ffe3-ce2b-46e3-ae71-6509e1613068
@openapiResponse 200 application/json {}
*/
func (h *Handlers) DeleteResponse(c echo.Context) error {
	uuid := c.Param("uuid")
	slog.Debug("Delete response", "uuid", uuid)
	err := h.db.DeleteResponse(c.Request().Context(), uuid)
	if err != nil {
		slog.Error("Delete response error", "error", err, "uuid", uuid)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

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
	slog.Debug("Clean test", "test_id", testID)
	err := h.db.Clean(c.Request().Context(), testID)
	if err != nil {
		slog.Error("Clean error", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, echo.Map{})
}

func (h *Handlers) Catch(c echo.Context) error {
	method := c.Request().Method
	path := c.Request().URL.Path

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

		slog.Debug("Catch request", "method", method, "path", path)
		err = h.db.SaveRequest(c.Request().Context(), request)
		if err != nil {
			slog.Error("Save request error", "error", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}
	}

	c.Response().Status = int(response.Status)
	for k, v := range response.Headers {
		c.Response().Header().Set(k, v)
	}
	_, err = c.Response().Write([]byte(response.Body))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return nil
}
