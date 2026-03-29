package handlers_test

import (
	"net/http"
	"testing"

	"github.com/zachatrocity/voyage/internal/api/handlers"
	"github.com/zachatrocity/voyage/internal/testutil"
)

func TestHealthCheck(t *testing.T) {
	tests := []struct {
		name       string
		wantStatus int
		wantKeys   map[string]string
	}{
		{
			name:       "returns 200 with status up",
			wantStatus: http.StatusOK,
			wantKeys: map[string]string{
				"status":  "up",
				"version": "0.1.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := testutil.NewTestEcho()
			e.GET("/health", handlers.HealthCheck)

			req, rec := testutil.NewRequest(http.MethodGet, "/health", nil)
			e.ServeHTTP(rec, req)

			testutil.AssertStatus(t, rec, tt.wantStatus)
			for key, val := range tt.wantKeys {
				testutil.AssertJSONKey(t, rec, key, val)
			}
		})
	}
}
