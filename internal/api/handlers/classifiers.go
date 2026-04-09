package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/zachatrocity/voyage/internal/classifier"
)

// classifiersResponse is the JSON shape for classifier config endpoints.
type classifiersResponse struct {
	Categories map[string]classifierCategoryJSON `json:"categories"`
}

type classifierCategoryJSON struct {
	Domains         []string `json:"domains"`
	SubjectKeywords []string `json:"subject_keywords"`
}

// toResponse converts a ClassifiersConfig to the JSON response shape.
func toClassifiersResponse(cfg classifier.ClassifiersConfig) classifiersResponse {
	resp := classifiersResponse{
		Categories: make(map[string]classifierCategoryJSON, len(cfg.Categories)),
	}
	for name, rule := range cfg.Categories {
		resp.Categories[name] = classifierCategoryJSON{
			Domains:         rule.Domains,
			SubjectKeywords: rule.SubjectKeywords,
		}
	}
	return resp
}

// fromRequest converts a classifiersResponse (from request body) to a ClassifiersConfig.
func fromClassifiersRequest(req classifiersResponse) classifier.ClassifiersConfig {
	cfg := classifier.ClassifiersConfig{
		Categories: make(map[string]classifier.CategoryRule, len(req.Categories)),
	}
	for name, cat := range req.Categories {
		cfg.Categories[name] = classifier.CategoryRule{
			Domains:         cat.Domains,
			SubjectKeywords: cat.SubjectKeywords,
		}
	}
	return cfg
}

// validateClassifiersRequest validates the request body and returns a list of errors.
func validateClassifiersRequest(req classifiersResponse) []string {
	var errs []string

	if len(req.Categories) == 0 {
		return []string{"at least one category must be present"}
	}

	for name, cat := range req.Categories {
		prefix := fmt.Sprintf("category %q", name)

		if strings.TrimSpace(name) == "" {
			errs = append(errs, "category names must be non-empty strings")
			continue
		}

		for i, d := range cat.Domains {
			if strings.TrimSpace(d) == "" {
				errs = append(errs, fmt.Sprintf("%s: domain at index %d must be a non-empty string", prefix, i))
			}
		}

		for i, kw := range cat.SubjectKeywords {
			if strings.TrimSpace(kw) == "" {
				errs = append(errs, fmt.Sprintf("%s: subject_keyword at index %d must be a non-empty string", prefix, i))
			}
		}
	}

	return errs
}

// GetClassifiers godoc
// @Summary Get classifier config
// @Description Returns the current classifier configuration (categories, domains, keywords)
// @Tags classifiers
// @Accept json
// @Produce json
// @Success 200 {object} classifiersResponse
// @Router /classifiers [get]
func GetClassifiers(c echo.Context) error {
	cfg := classifier.GetConfig()
	return c.JSON(http.StatusOK, toClassifiersResponse(cfg))
}

// UpdateClassifiers godoc
// @Summary Update classifier config
// @Description Replaces the classifier configuration. Validates input and persists to YAML file.
// @Tags classifiers
// @Accept json
// @Produce json
// @Param body body classifiersResponse true "Classifier config"
// @Success 200 {object} classifiersResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /classifiers [put]
func UpdateClassifiers(c echo.Context) error {
	var req classifiersResponse
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body: " + err.Error(),
		})
	}

	// Validate
	if errs := validateClassifiersRequest(req); len(errs) > 0 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":  "validation failed",
			"errors": errs,
		})
	}

	// Check we have a file path to write to
	path := classifier.GetConfigPath()
	if path == "" {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "classifiers file path not configured; set VOYAGE_CLASSIFIERS env var",
		})
	}

	// Convert and write
	cfg := fromClassifiersRequest(req)
	if err := classifier.WriteConfig(path, cfg); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to write classifiers file: " + err.Error(),
		})
	}

	// Reload in-memory config
	if err := classifier.ReloadConfig(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to reload classifiers config: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, toClassifiersResponse(classifier.GetConfig()))
}

// ResetClassifiers godoc
// @Summary Reset classifier config to defaults
// @Description Restores the built-in default classifier configuration and persists it to the YAML file.
// @Tags classifiers
// @Accept json
// @Produce json
// @Success 200 {object} classifiersResponse
// @Failure 500 {object} map[string]string
// @Router /classifiers/reset [post]
func ResetClassifiers(c echo.Context) error {
	path := classifier.GetConfigPath()
	if path == "" {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "classifiers file path not configured; set VOYAGE_CLASSIFIERS env var",
		})
	}

	if err := classifier.WriteDefaultConfig(path); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to write default classifiers file: " + err.Error(),
		})
	}

	if err := classifier.ReloadConfig(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to reload classifiers config: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, toClassifiersResponse(classifier.GetConfig()))
}
