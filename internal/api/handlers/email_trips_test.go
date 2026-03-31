package handlers_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/zachatrocity/voyage/internal/api/handlers"
	"github.com/zachatrocity/voyage/internal/testutil"
	"github.com/zachatrocity/voyage/internal/trips"
)

func TestAssociateTripEmail_TripNotFound(t *testing.T) {
	h := newTripHandler(t)
	e := testutil.NewTestEcho()
	e.POST("/api/v1/email/:id/trip/:tripId", h.AssociateTripEmail)

	req, rec := testutil.NewRequest(http.MethodPost, "/api/v1/email/abc123/trip/nonexistent-trip", nil)
	e.ServeHTTP(rec, req)

	testutil.AssertStatus(t, rec, http.StatusNotFound)
	testutil.AssertJSONKey(t, rec, "error", "trip not found")
}

func TestAssociateTripEmail_RouteParams(t *testing.T) {
	// Create a trip so the trip-exists check passes; the notmuch call will
	// fail because there's no DB, but we can verify the route params are
	// parsed by checking we get past the 404 trip-not-found check.
	store, err := trips.NewJSONStore(t.TempDir() + "/trips.json")
	if err != nil {
		t.Fatalf("NewJSONStore: %v", err)
	}
	svc := trips.NewService(store)
	svc.CreateTrip("Test Trip", "")
	h := handlers.NewTripHandler(svc)

	e := testutil.NewTestEcho()
	e.POST("/api/v1/email/:id/trip/:tripId", h.AssociateTripEmail)

	req, rec := testutil.NewRequest(http.MethodPost, "/api/v1/email/msg-123/trip/test-trip", nil)
	e.ServeHTTP(rec, req)

	// We expect a 500 (notmuch DB not available) rather than 404,
	// proving the route params were parsed correctly and the trip was found.
	if rec.Code == http.StatusNotFound {
		var body map[string]string
		json.Unmarshal(rec.Body.Bytes(), &body)
		if body["error"] == "trip not found" {
			t.Error("got trip not found; route param :tripId was not parsed correctly")
		}
	}
}

func TestListTripEmails_TripNotFound(t *testing.T) {
	h := newTripHandler(t)
	e := testutil.NewTestEcho()
	e.GET("/api/v1/trips/:id/emails", h.ListTripEmails)

	req, rec := testutil.NewRequest(http.MethodGet, "/api/v1/trips/nonexistent/emails", nil)
	e.ServeHTTP(rec, req)

	testutil.AssertStatus(t, rec, http.StatusNotFound)
	testutil.AssertJSONKey(t, rec, "error", "trip not found")
}

func TestListTripEmails_RouteParams(t *testing.T) {
	// Create a trip so we get past the 404 check
	store, err := trips.NewJSONStore(t.TempDir() + "/trips.json")
	if err != nil {
		t.Fatalf("NewJSONStore: %v", err)
	}
	svc := trips.NewService(store)
	svc.CreateTrip("Test Trip", "")
	h := handlers.NewTripHandler(svc)

	e := testutil.NewTestEcho()
	e.GET("/api/v1/trips/:id/emails", h.ListTripEmails)

	req, rec := testutil.NewRequest(http.MethodGet, "/api/v1/trips/test-trip/emails", nil)
	e.ServeHTTP(rec, req)

	// Should not be 404 "trip not found" — the trip exists, so we get past
	// the validation and into notmuch (which will 500 without a DB).
	if rec.Code == http.StatusNotFound {
		var body map[string]string
		json.Unmarshal(rec.Body.Bytes(), &body)
		if body["error"] == "trip not found" {
			t.Error("got trip not found; route param :id was not parsed correctly")
		}
	}
}
