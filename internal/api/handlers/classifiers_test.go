package handlers_test

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/zachatrocity/voyage/internal/api/handlers"
	"github.com/zachatrocity/voyage/internal/classifier"
	"github.com/zachatrocity/voyage/internal/testutil"
)

// setupClassifiers creates a temp file for classifiers, calls SetConfigPath with the
// default config written to it, and registers routes on a test Echo instance.
func setupClassifiers(t *testing.T) *echo.Echo {
	t.Helper()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "classifiers.yaml")

	// Write defaults so the file exists.
	if err := classifier.WriteDefaultConfig(cfgPath); err != nil {
		t.Fatalf("WriteDefaultConfig: %v", err)
	}
	if err := classifier.SetConfigPath(cfgPath); err != nil {
		t.Fatalf("SetConfigPath: %v", err)
	}

	e := testutil.NewTestEcho()
	e.GET("/api/v1/classifiers", handlers.GetClassifiers)
	e.PUT("/api/v1/classifiers", handlers.UpdateClassifiers)
	e.POST("/api/v1/classifiers/reset", handlers.ResetClassifiers)
	return e
}

// --- GET /api/v1/classifiers ---

func TestGetClassifiers_ReturnsCurrentConfig(t *testing.T) {
	e := setupClassifiers(t)

	req, rec := testutil.NewRequest(http.MethodGet, "/api/v1/classifiers", nil)
	e.ServeHTTP(rec, req)

	testutil.AssertStatus(t, rec, http.StatusOK)

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	cats, ok := body["categories"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected categories map in response, got: %v", body)
	}
	if _, exists := cats["flight"]; !exists {
		t.Error("expected 'flight' category in default config")
	}
}

// --- PUT /api/v1/classifiers ---

func TestUpdateClassifiers_HappyPath(t *testing.T) {
	e := setupClassifiers(t)

	payload := map[string]interface{}{
		"categories": map[string]interface{}{
			"custom": map[string]interface{}{
				"domains":          []string{"myairline.com"},
				"subject_keywords": []string{"my flight"},
			},
		},
	}

	req, rec := testutil.NewRequest(http.MethodPut, "/api/v1/classifiers", payload)
	e.ServeHTTP(rec, req)

	testutil.AssertStatus(t, rec, http.StatusOK)

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	cats := body["categories"].(map[string]interface{})
	if _, ok := cats["custom"]; !ok {
		t.Error("expected 'custom' category in response after update")
	}
}

func TestUpdateClassifiers_ValidationError_EmptyCategories(t *testing.T) {
	e := setupClassifiers(t)

	payload := map[string]interface{}{
		"categories": map[string]interface{}{},
	}

	req, rec := testutil.NewRequest(http.MethodPut, "/api/v1/classifiers", payload)
	e.ServeHTTP(rec, req)

	testutil.AssertStatus(t, rec, http.StatusBadRequest)

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["error"] == nil {
		t.Error("expected error field in 400 response")
	}
}

func TestUpdateClassifiers_ValidationError_EmptyDomain(t *testing.T) {
	e := setupClassifiers(t)

	payload := map[string]interface{}{
		"categories": map[string]interface{}{
			"flight": map[string]interface{}{
				"domains":          []string{""},
				"subject_keywords": []string{"flight"},
			},
		},
	}

	req, rec := testutil.NewRequest(http.MethodPut, "/api/v1/classifiers", payload)
	e.ServeHTTP(rec, req)

	testutil.AssertStatus(t, rec, http.StatusBadRequest)

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	errList, ok := body["errors"].([]interface{})
	if !ok || len(errList) == 0 {
		t.Error("expected errors array with at least one entry")
	}
}

func TestUpdateClassifiers_ValidationError_EmptyKeyword(t *testing.T) {
	e := setupClassifiers(t)

	payload := map[string]interface{}{
		"categories": map[string]interface{}{
			"flight": map[string]interface{}{
				"domains":          []string{"delta.com"},
				"subject_keywords": []string{""},
			},
		},
	}

	req, rec := testutil.NewRequest(http.MethodPut, "/api/v1/classifiers", payload)
	e.ServeHTTP(rec, req)

	testutil.AssertStatus(t, rec, http.StatusBadRequest)
}

func TestUpdateClassifiers_PersistsAndReloads(t *testing.T) {
	e := setupClassifiers(t)

	// Update to a custom config.
	payload := map[string]interface{}{
		"categories": map[string]interface{}{
			"hotel": map[string]interface{}{
				"domains":          []string{"myhotel.com"},
				"subject_keywords": []string{"stay"},
			},
		},
	}
	req, rec := testutil.NewRequest(http.MethodPut, "/api/v1/classifiers", payload)
	e.ServeHTTP(rec, req)
	testutil.AssertStatus(t, rec, http.StatusOK)

	// Now GET should return the updated config.
	req2, rec2 := testutil.NewRequest(http.MethodGet, "/api/v1/classifiers", nil)
	e.ServeHTTP(rec2, req2)
	testutil.AssertStatus(t, rec2, http.StatusOK)

	var body map[string]interface{}
	if err := json.Unmarshal(rec2.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	cats := body["categories"].(map[string]interface{})
	if _, ok := cats["hotel"]; !ok {
		t.Error("expected 'hotel' category after update")
	}
	if _, ok := cats["flight"]; ok {
		t.Error("'flight' category should be gone after full replacement")
	}
}

// --- POST /api/v1/classifiers/reset ---

func TestResetClassifiers_RestoresDefaults(t *testing.T) {
	e := setupClassifiers(t)

	// First, replace with minimal config.
	payload := map[string]interface{}{
		"categories": map[string]interface{}{
			"custom": map[string]interface{}{
				"domains":          []string{"custom.com"},
				"subject_keywords": []string{"custom"},
			},
		},
	}
	req, rec := testutil.NewRequest(http.MethodPut, "/api/v1/classifiers", payload)
	e.ServeHTTP(rec, req)
	testutil.AssertStatus(t, rec, http.StatusOK)

	// Reset.
	req2, rec2 := testutil.NewRequest(http.MethodPost, "/api/v1/classifiers/reset", nil)
	e.ServeHTTP(rec2, req2)
	testutil.AssertStatus(t, rec2, http.StatusOK)

	var body map[string]interface{}
	if err := json.Unmarshal(rec2.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	cats := body["categories"].(map[string]interface{})

	// Default categories should be present.
	for _, expected := range []string{"flight", "hotel", "cruise", "car_rental", "activity"} {
		if _, ok := cats[expected]; !ok {
			t.Errorf("expected default category %q after reset", expected)
		}
	}
	// Custom category should be gone.
	if _, ok := cats["custom"]; ok {
		t.Error("'custom' category should be gone after reset")
	}
}
