package handlers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/zachatrocity/voyage/internal/notmuch"
)

// HealthCheck provides a simple health check endpoint
func HealthCheck(c echo.Context) error {
	// Check if notmuch database is accessible
	dbStatus := "ok"
	if err := notmuch.CheckDatabaseConnection(); err != nil {
		dbStatus = "error: " + err.Error()
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "up",
		"database":  dbStatus,
		"version":   "0.1.0",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// Search handles simple search queries against the notmuch database
func Search(c echo.Context) error {
	// Get query parameter
	query := c.QueryParam("q")
	if query == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Query parameter 'q' is required",
		})
	}

	// Get optional limit parameter
	limit := c.QueryParam("limit")
	if limit == "" {
		limit = "50"
	}

	// Perform search
	results, err := notmuch.Search(query, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to search emails: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, results)
}

// GetEmail retrieves a single email by its message ID
func GetEmail(c echo.Context) error {
	// Get message ID from URL parameter
	messageID := c.Param("id")
	if messageID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Message ID is required",
		})
	}

	// Get email details
	email, err := notmuch.GetEmail(messageID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve email: " + err.Error(),
		})
	}

	// Check if email was found
	if email == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Email not found",
		})
	}

	return c.JSON(http.StatusOK, email)
}
