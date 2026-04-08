package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	voyagenotmuch "github.com/zachatrocity/voyage/internal/notmuch"
)

func respondNotmuchError(c echo.Context, baseMessage string, err error) error {
	payload := map[string]any{
		"error": baseMessage + ": " + err.Error(),
	}

	if opErr, ok := voyagenotmuch.AsOperationError(err); ok {
		payload["code"] = opErr.Code
		payload["retryable"] = opErr.Retryable
		payload["details"] = opErr.Details()
	}

	return c.JSON(http.StatusInternalServerError, payload)
}
