package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/zachatrocity/voyage/internal/classifier"
)

func setupClassifierTestConfig(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	classifier.SetConfigPath(filepath.Join(dir, "classifiers.yaml"))
	if _, err := classifier.WriteDefaultConfig(); err != nil {
		t.Fatalf("WriteDefaultConfig: %v", err)
	}
}

func TestGetClassifiers(t *testing.T) {
	setupClassifierTestConfig(t)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/classifiers", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := GetClassifiers(c); err != nil {
		t.Fatalf("GetClassifiers error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}

	var cfg classifier.ClassifiersConfig
	if err := json.Unmarshal(rec.Body.Bytes(), &cfg); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(cfg.Categories) == 0 {
		t.Fatal("expected categories in response")
	}
}

func TestUpdateClassifiers(t *testing.T) {
	setupClassifierTestConfig(t)
	e := echo.New()
	body := `{"categories":{"flight":{"domains":["myair.com"],"subject_keywords":["boarding pass"]}}}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/classifiers", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := UpdateClassifiers(c); err != nil {
		t.Fatalf("UpdateClassifiers error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d, body=%s", rec.Code, rec.Body.String())
	}

	cfg := classifier.GetConfig()
	if got := cfg.Categories["flight"].Domains[0]; got != "myair.com" {
		t.Fatalf("expected updated domain myair.com, got %q", got)
	}
}

func TestUpdateClassifiersValidationError(t *testing.T) {
	setupClassifierTestConfig(t)
	e := echo.New()
	body := `{"categories":{"flight":{"domains":[""],"subject_keywords":[]}}}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/classifiers", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := UpdateClassifiers(c); err != nil {
		t.Fatalf("UpdateClassifiers error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", rec.Code)
	}
}

func TestResetClassifiers(t *testing.T) {
	setupClassifierTestConfig(t)
	if err := classifier.WriteConfig(classifier.ClassifiersConfig{Categories: map[string]classifier.CategoryRule{"flight": {Domains: []string{"myair.com"}}}}); err != nil {
		t.Fatalf("seed custom config: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/classifiers/reset", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := ResetClassifiers(c); err != nil {
		t.Fatalf("ResetClassifiers error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}

	cfg := classifier.GetConfig()
	if _, ok := cfg.Categories["hotel"]; !ok {
		t.Fatal("expected reset defaults to include hotel category")
	}
}
