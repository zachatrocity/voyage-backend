package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/zachatrocity/voyage/internal/classifier"
)

// GetClassifiers godoc
// @Summary Get classifier config
// @Description Returns the current classifier configuration (categories, domains, keywords)
// @Tags classifiers
// @Accept json
// @Produce json
// @Success 200 {object} classifier.ClassifiersConfig
// @Router /classifiers [get]
func GetClassifiers(c echo.Context) error {
	return c.JSON(http.StatusOK, classifier.GetConfig())
}

// UpdateClassifiers godoc
// @Summary Update classifier config
// @Description Replaces the classifier configuration. Validates input and persists to YAML file.
// @Tags classifiers
// @Accept json
// @Produce json
// @Param body body classifier.ClassifiersConfig true "Classifier config"
// @Success 200 {object} classifier.ClassifiersConfig
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /classifiers [put]
func UpdateClassifiers(c echo.Context) error {
	var req classifier.ClassifiersConfig
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "invalid request body",
		})
	}

	if err := classifier.ValidateConfig(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
	}

	if err := classifier.WriteConfig(req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, classifier.GetConfig())
}

// ResetClassifiers godoc
// @Summary Reset classifier config to defaults
// @Description Restores the built-in default classifier configuration and persists it to the YAML file.
// @Tags classifiers
// @Accept json
// @Produce json
// @Success 200 {object} classifier.ClassifiersConfig
// @Failure 500 {object} map[string]string
// @Router /classifiers/reset [post]
func ResetClassifiers(c echo.Context) error {
	cfg, err := classifier.WriteDefaultConfig()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, cfg)
}
