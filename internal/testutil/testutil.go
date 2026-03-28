package testutil

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// NewTestEcho returns an Echo instance configured the same way as production
// (logger, recover, CORS) but without API key auth or doc routes.
func NewTestEcho() *echo.Echo {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderContentType, echo.HeaderAccept, "X-API-Key"},
	}))
	return e
}

// NewRequest creates an *http.Request and a *httptest.ResponseRecorder.
// If body is non-nil it is JSON-encoded and Content-Type is set to application/json.
func NewRequest(method, path string, body interface{}) (*http.Request, *httptest.ResponseRecorder) {
	var req *http.Request
	if body != nil {
		b, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rec := httptest.NewRecorder()
	return req, rec
}

// AssertStatus fails the test if the recorded status code does not match want.
func AssertStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if got := rec.Code; got != want {
		t.Errorf("status = %d; want %d\nbody: %s", got, want, rec.Body.String())
	}
}

// AssertJSONKey decodes the response body as JSON and asserts that the
// top-level string value at key equals wantVal.
func AssertJSONKey(t *testing.T, rec *httptest.ResponseRecorder, key, wantVal string) {
	t.Helper()
	var m map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("failed to decode JSON body: %v\nbody: %s", err, rec.Body.String())
	}
	got, ok := m[key]
	if !ok {
		t.Fatalf("key %q not found in JSON response; keys: %v", key, keys(m))
	}
	gotStr, ok := got.(string)
	if !ok {
		t.Fatalf("key %q is not a string; got %T: %v", key, got, got)
	}
	if gotStr != wantVal {
		t.Errorf("JSON[%q] = %q; want %q", key, gotStr, wantVal)
	}
}

func keys(m map[string]interface{}) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}
