package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/zachatrocity/voyage/internal/notmuch"
)

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Get the health status of the API and database connection
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
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

// Search godoc
// @Summary Search emails
// @Description Search for emails using notmuch query
// @Tags search
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param limit query string false "Result limit" default(50)
// @Param sort query string false "Sort order (oldest_first, newest_first)" default(newest_first)
// @Success 200 {object} notmuch.SearchResults
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /search [get]
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

	// Get optional sort parameter
	sortParam := c.QueryParam("sort")
	var sortType notmuch.SortType
	switch sortParam {
	case "oldest_first":
		sortType = notmuch.SortOldestFirst
	default:
		sortType = notmuch.SortNewestFirst // Default to newest first
	}

	// Log the sort parameter for debugging
	log.Printf("Search request with query: %s, sort param: %s, sort type: %d", query, sortParam, sortType)

	// Perform search
	results, err := notmuch.Search(query, limit, sortType)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to search emails: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, results)
}

// GetEmail godoc
// @Summary Get email by ID
// @Description Retrieve a single email by its message ID
// @Tags emails
// @Accept json
// @Produce json
// @Param id path string true "Message ID"
// @Success 200 {object} notmuch.EmailResult
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /emails/{id} [get]
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

// TagEmail godoc
// @Summary Tag an email
// @Description Add a tag to an email by its message ID
// @Tags emails
// @Accept json
// @Produce json
// @Param id path string true "Message ID"
// @Param tag path string true "Tag to add"
// @Success 200 {object} notmuch.EmailResult
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /emails/{id}/tags/{tag} [post]
func TagEmail(c echo.Context) error {
	// Get message ID from URL parameter
	messageID := c.Param("id")
	if messageID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Message ID is required",
		})
	}

	// Get message tag from URL parameter
	tag := c.Param("tag")
	if tag == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "tag is required",
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

	taggedEmail, err := notmuch.TagEmail(messageID, tag)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to tag email: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, taggedEmail)
}
