package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/onrik/supermock/pkg/models"
)

var (
	_ = models.Email{}
)

/*
Emails
@openapi GET /_emails
@openapiSummary Get email messages
@openapiResponse 501 application/json {}
@openapiResponse 200 application/json {"emails": []models.Email}
*/
func (h *Handlers) Emails(c echo.Context) error {
	if h.smtp == nil {
		return echo.ErrNotImplemented
	}

	return c.JSON(http.StatusOK, echo.Map{
		"emails": h.smtp.Emails(),
	})
}

/*
Emails
@openapi DELETE /_emails
@openapiSummary Delete all email messages
@openapiResponse 501 application/json {}
@openapiResponse 200 application/json {}
*/
func (h *Handlers) EmailsDelete(c echo.Context) error {
	if h.smtp == nil {
		return echo.ErrNotImplemented
	}

	h.smtp.Purge()

	return c.JSON(http.StatusOK, echo.Map{})
}
