package handlers

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/zachatrocity/voyage/internal/trips"
)

// TripHandler holds trip-related HTTP handlers.
type TripHandler struct {
	svc *trips.Service
}

// NewTripHandler creates a TripHandler backed by the given service.
func NewTripHandler(svc *trips.Service) *TripHandler {
	return &TripHandler{svc: svc}
}

// tripResponse is the JSON shape returned for a single trip.
type tripResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	DateRange      string `json:"date_range"`
	EmailCount     int    `json:"email_count"`
	ConfirmedCount int    `json:"confirmed_count"`
	CreatedAt      string `json:"created_at"`
}

func toTripResponse(t *trips.Trip) tripResponse {
	return tripResponse{
		ID:             t.ID,
		Name:           t.Name,
		DateRange:      t.DateRange,
		EmailCount:     0,
		ConfirmedCount: 0,
		CreatedAt:      t.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ListTrips godoc
// @Summary List all trips
// @Description Get a list of all trips
// @Tags trips
// @Accept json
// @Produce json
// @Success 200 {object} map[string][]tripResponse
// @Failure 500 {object} map[string]string
// @Router /trips [get]
func (h *TripHandler) ListTrips(c echo.Context) error {
	list, err := h.svc.ListTrips()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to list trips: " + err.Error(),
		})
	}
	resp := make([]tripResponse, len(list))
	for i, t := range list {
		resp[i] = toTripResponse(t)
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"trips": resp,
	})
}

// createTripRequest is the expected JSON body for creating a trip.
type createTripRequest struct {
	Name      string `json:"name"`
	DateRange string `json:"date_range"`
}

// CreateTrip godoc
// @Summary Create a new trip
// @Description Create a trip with a name and date range
// @Tags trips
// @Accept json
// @Produce json
// @Param trip body createTripRequest true "Trip to create"
// @Success 201 {object} tripResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /trips [post]
func (h *TripHandler) CreateTrip(c echo.Context) error {
	var req createTripRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}
	if strings.TrimSpace(req.Name) == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "name is required",
		})
	}

	t, err := h.svc.CreateTrip(req.Name, req.DateRange)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create trip: " + err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, toTripResponse(t))
}

// DeleteTrip godoc
// @Summary Delete a trip
// @Description Delete a trip by id
// @Tags trips
// @Accept json
// @Produce json
// @Param id path string true "Trip ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /trips/{id} [delete]
func (h *TripHandler) DeleteTrip(c echo.Context) error {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "id is required",
		})
	}

	if _, err := h.svc.GetTrip(id); err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "trip not found",
		})
	}

	if err := h.svc.DeleteTrip(id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete trip: " + err.Error(),
		})
	}

	return c.NoContent(http.StatusNoContent)
}
