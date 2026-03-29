package handlers_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/zachatrocity/voyage/internal/api/handlers"
	"github.com/zachatrocity/voyage/internal/testutil"
	"github.com/zachatrocity/voyage/internal/trips"
)

func newTripHandler(t *testing.T) *handlers.TripHandler {
	t.Helper()
	store, err := trips.NewJSONStore(t.TempDir() + "/trips.json")
	if err != nil {
		t.Fatalf("NewJSONStore: %v", err)
	}
	return handlers.NewTripHandler(trips.NewService(store))
}

func TestListTrips_Empty(t *testing.T) {
	h := newTripHandler(t)
	e := testutil.NewTestEcho()
	e.GET("/api/v1/trips", h.ListTrips)

	req, rec := testutil.NewRequest(http.MethodGet, "/api/v1/trips", nil)
	e.ServeHTTP(rec, req)

	testutil.AssertStatus(t, rec, http.StatusOK)

	var body map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if string(body["trips"]) != "[]" {
		t.Errorf("expected empty trips array, got %s", body["trips"])
	}
}

func TestCreateTrip(t *testing.T) {
	h := newTripHandler(t)
	e := testutil.NewTestEcho()
	e.POST("/api/v1/trips", h.CreateTrip)

	req, rec := testutil.NewRequest(http.MethodPost, "/api/v1/trips", map[string]string{
		"name":       "2026 Caribbean Cruise",
		"date_range": "Jun 18 – Jun 25, 2026",
	})
	e.ServeHTTP(rec, req)

	testutil.AssertStatus(t, rec, http.StatusCreated)

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["id"] != "2026-caribbean-cruise" {
		t.Errorf("id = %v; want 2026-caribbean-cruise", body["id"])
	}
	if body["name"] != "2026 Caribbean Cruise" {
		t.Errorf("name = %v; want 2026 Caribbean Cruise", body["name"])
	}
}

func TestCreateTrip_MissingName(t *testing.T) {
	h := newTripHandler(t)
	e := testutil.NewTestEcho()
	e.POST("/api/v1/trips", h.CreateTrip)

	req, rec := testutil.NewRequest(http.MethodPost, "/api/v1/trips", map[string]string{
		"date_range": "Jun 18 – Jun 25, 2026",
	})
	e.ServeHTTP(rec, req)

	testutil.AssertStatus(t, rec, http.StatusBadRequest)
}

func TestCreateTrip_Duplicate(t *testing.T) {
	h := newTripHandler(t)
	e := testutil.NewTestEcho()
	e.POST("/api/v1/trips", h.CreateTrip)

	body := map[string]string{
		"name":       "2026 Caribbean Cruise",
		"date_range": "Jun 18 – Jun 25, 2026",
	}

	// First create succeeds
	req, rec := testutil.NewRequest(http.MethodPost, "/api/v1/trips", body)
	e.ServeHTTP(rec, req)
	testutil.AssertStatus(t, rec, http.StatusCreated)

	// Second create returns 409
	req, rec = testutil.NewRequest(http.MethodPost, "/api/v1/trips", body)
	e.ServeHTTP(rec, req)
	testutil.AssertStatus(t, rec, http.StatusConflict)
}

func TestListTrips_AfterCreate(t *testing.T) {
	h := newTripHandler(t)
	e := testutil.NewTestEcho()
	e.POST("/api/v1/trips", h.CreateTrip)
	e.GET("/api/v1/trips", h.ListTrips)

	// Create a trip
	req, rec := testutil.NewRequest(http.MethodPost, "/api/v1/trips", map[string]string{
		"name":       "2026 Caribbean Cruise",
		"date_range": "Jun 18 – Jun 25, 2026",
	})
	e.ServeHTTP(rec, req)
	testutil.AssertStatus(t, rec, http.StatusCreated)

	// List trips
	req, rec = testutil.NewRequest(http.MethodGet, "/api/v1/trips", nil)
	e.ServeHTTP(rec, req)
	testutil.AssertStatus(t, rec, http.StatusOK)

	var body struct {
		Trips []map[string]interface{} `json:"trips"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Trips) != 1 {
		t.Fatalf("expected 1 trip, got %d", len(body.Trips))
	}
	if body.Trips[0]["id"] != "2026-caribbean-cruise" {
		t.Errorf("trip id = %v; want 2026-caribbean-cruise", body.Trips[0]["id"])
	}
}
