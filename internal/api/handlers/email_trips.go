package handlers

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	voyagenotmuch "github.com/zachatrocity/voyage/internal/notmuch"
)

// associateTripEmailResponse is the JSON shape returned when associating an email with a trip.
type associateTripEmailResponse struct {
	MessageID string   `json:"message_id" example:"<abc123@gmail.com>"`
	TripID    string   `json:"trip_id" example:"2026-cruise"`
	Tags      []string `json:"tags" example:"inbox,trip:2026-cruise,travel"`
}

// TripEmailItem mirrors notmuch.EmailResult for Swagger doc generation.
// @Description An email associated with a trip
type TripEmailItem struct {
	MessageID string   `json:"message_id"`
	ThreadID  string   `json:"thread_id"`
	Date      string   `json:"date"`
	From      string   `json:"from"`
	Subject   string   `json:"subject"`
	Tags      []string `json:"tags"`
	Filename  string   `json:"filename"`
}

// tripEmailsResponse is the JSON shape returned when listing emails for a trip.
type tripEmailsResponse struct {
	TripID string          `json:"trip_id" example:"2026-cruise"`
	Emails []TripEmailItem `json:"emails"`
}

// AssociateTripEmail godoc
// @Summary Associate an email with a trip
// @Description Remove any existing trip:* tags and apply trip:{tripId} to the email
// @Tags email,trips
// @Accept json
// @Produce json
// @Param id path string true "Email message ID"
// @Param tripId path string true "Trip ID"
// @Success 200 {object} associateTripEmailResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /email/{id}/trip/{tripId} [post]
func (h *TripHandler) AssociateTripEmail(c echo.Context) error {
	messageID := c.Param("id")
	tripID := c.Param("tripId")

	// Validate trip exists
	_, err := h.svc.GetTrip(tripID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "trip not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to look up trip: " + err.Error(),
		})
	}

	// Replace trip tags on the email
	result, err := voyagenotmuch.ReplaceTripTag(messageID, tripID)
	if err != nil {
		return respondNotmuchError(c, "Failed to update email tags", err)
	}
	if result == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "email not found",
		})
	}

	return c.JSON(http.StatusOK, associateTripEmailResponse{
		MessageID: result.MessageID,
		TripID:    tripID,
		Tags:      result.Tags,
	})
}

// ListTripEmails godoc
// @Summary List emails for a trip
// @Description Get all emails tagged with a specific trip, sorted chronologically
// @Tags email,trips
// @Accept json
// @Produce json
// @Param id path string true "Trip ID"
// @Success 200 {object} tripEmailsResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /trips/{id}/emails [get]
func (h *TripHandler) ListTripEmails(c echo.Context) error {
	tripID := c.Param("id")

	// Validate trip exists
	_, err := h.svc.GetTrip(tripID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "trip not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to look up trip: " + err.Error(),
		})
	}

	// Search for emails tagged with this trip
	results, err := voyagenotmuch.Search("tag:trip:"+tripID, "500", voyagenotmuch.SortOldestFirst)
	if err != nil {
		return respondNotmuchError(c, "Failed to search emails", err)
	}

	items := make([]TripEmailItem, 0, len(results.Results))
	for _, e := range results.Results {
		items = append(items, TripEmailItem{
			MessageID: e.MessageID,
			ThreadID:  e.ThreadID,
			Date:      e.Date.Format("2006-01-02T15:04:05Z07:00"),
			From:      e.From,
			Subject:   e.Subject,
			Tags:      e.Tags,
			Filename:  e.Filename,
		})
	}

	return c.JSON(http.StatusOK, tripEmailsResponse{
		TripID: tripID,
		Emails: items,
	})
}
