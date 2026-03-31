package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/zachatrocity/voyage/internal/sync"
)

// SyncHandler holds dependencies for sync endpoints.
type SyncHandler struct {
	manager *sync.SyncManager
}

// NewSyncHandler creates a new SyncHandler.
func NewSyncHandler(manager *sync.SyncManager) *SyncHandler {
	return &SyncHandler{manager: manager}
}

// TriggerSync godoc
// @Summary Trigger mail sync
// @Description Start a mail sync process in the background
// @Tags sync
// @Accept json
// @Produce json
// @Success 202 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /sync [post]
func (h *SyncHandler) TriggerSync(c echo.Context) error {
	if err := h.manager.TriggerSync(); err != nil {
		return c.JSON(http.StatusConflict, map[string]string{
			"status":  "error",
			"message": "Sync is already running",
		})
	}
	return c.JSON(http.StatusAccepted, map[string]string{
		"status":  "started",
		"message": "Mail sync started in background",
	})
}

// GetSyncStatus godoc
// @Summary Get sync status
// @Description Get the current status of the mail sync process
// @Tags sync
// @Accept json
// @Produce json
// @Success 200 {object} sync.SyncStatus
// @Router /sync/status [get]
func (h *SyncHandler) GetSyncStatus(c echo.Context) error {
	return c.JSON(http.StatusOK, h.manager.Status())
}
