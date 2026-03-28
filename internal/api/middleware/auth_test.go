package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestAPIKeyMiddleware(t *testing.T) {
	okHandler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	tests := []struct {
		name       string
		apiKey     string
		headerKey  string
		headerVal  string
		wantStatus int
	}{
		{
			name:       "auth disabled (empty key) — always passes",
			apiKey:     "",
			wantStatus: http.StatusOK,
		},
		{
			name:       "auth enabled, correct X-API-Key",
			apiKey:     "secret123",
			headerKey:  "X-API-Key",
			headerVal:  "secret123",
			wantStatus: http.StatusOK,
		},
		{
			name:       "auth enabled, wrong X-API-Key",
			apiKey:     "secret123",
			headerKey:  "X-API-Key",
			headerVal:  "wrong",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "auth enabled, missing key",
			apiKey:     "secret123",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "auth enabled, key via Authorization Bearer",
			apiKey:     "secret123",
			headerKey:  "Authorization",
			headerVal:  "Bearer secret123",
			wantStatus: http.StatusOK,
		},
		{
			name:       "auth enabled, wrong Authorization Bearer",
			apiKey:     "secret123",
			headerKey:  "Authorization",
			headerVal:  "Bearer wrong",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.headerKey != "" {
				req.Header.Set(tt.headerKey, tt.headerVal)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			h := APIKeyMiddleware(tt.apiKey)(okHandler)
			err := h(c)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if rec.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}
