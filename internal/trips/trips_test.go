package trips

import (
	"fmt"
	"sync"
	"testing"
)

// MemStore is a simple in-memory Store for testing.
type MemStore struct {
	mu    sync.Mutex
	trips map[string]*Trip
}

func NewMemStore() *MemStore {
	return &MemStore{trips: make(map[string]*Trip)}
}

func (m *MemStore) Create(t *Trip) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.trips[t.ID]; exists {
		return fmt.Errorf("trip %q already exists", t.ID)
	}
	m.trips[t.ID] = t
	return nil
}

func (m *MemStore) List() ([]*Trip, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]*Trip, 0, len(m.trips))
	for _, t := range m.trips {
		out = append(out, t)
	}
	return out, nil
}

func (m *MemStore) Get(id string) (*Trip, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.trips[id]
	if !ok {
		return nil, fmt.Errorf("trip %q not found", id)
	}
	return t, nil
}

func (m *MemStore) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.trips[id]; !ok {
		return fmt.Errorf("trip %q not found", id)
	}
	delete(m.trips, id)
	return nil
}

func TestCreateTrip(t *testing.T) {
	svc := NewService(NewMemStore())
	trip, err := svc.CreateTrip("2026 Alaska Cruise", "June 1–14, 2026")
	if err != nil {
		t.Fatalf("CreateTrip: %v", err)
	}
	if trip.ID != "2026-alaska-cruise" {
		t.Errorf("ID = %q, want %q", trip.ID, "2026-alaska-cruise")
	}
	if trip.Name != "2026 Alaska Cruise" {
		t.Errorf("Name = %q, want %q", trip.Name, "2026 Alaska Cruise")
	}
	if trip.DateRange != "June 1–14, 2026" {
		t.Errorf("DateRange = %q, want %q", trip.DateRange, "June 1–14, 2026")
	}
	if trip.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
}

func TestListTrips(t *testing.T) {
	svc := NewService(NewMemStore())
	if _, err := svc.CreateTrip("Trip A", "Jan"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateTrip("Trip B", "Feb"); err != nil {
		t.Fatal(err)
	}
	trips, err := svc.ListTrips()
	if err != nil {
		t.Fatalf("ListTrips: %v", err)
	}
	if len(trips) != 2 {
		t.Fatalf("got %d trips, want 2", len(trips))
	}
}

func TestGetTrip(t *testing.T) {
	svc := NewService(NewMemStore())
	created, err := svc.CreateTrip("My Trip", "March")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("existing", func(t *testing.T) {
		got, err := svc.GetTrip(created.ID)
		if err != nil {
			t.Fatalf("GetTrip: %v", err)
		}
		if got.Name != created.Name {
			t.Errorf("Name = %q, want %q", got.Name, created.Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.GetTrip("does-not-exist")
		if err == nil {
			t.Error("expected error for non-existent trip")
		}
	})
}

func TestDeleteTrip(t *testing.T) {
	svc := NewService(NewMemStore())
	created, err := svc.CreateTrip("Delete Me", "April")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("existing", func(t *testing.T) {
		if err := svc.DeleteTrip(created.ID); err != nil {
			t.Fatalf("DeleteTrip: %v", err)
		}
		_, err := svc.GetTrip(created.ID)
		if err == nil {
			t.Error("expected error after delete")
		}
	})

	t.Run("not found", func(t *testing.T) {
		err := svc.DeleteTrip("does-not-exist")
		if err == nil {
			t.Error("expected error for non-existent trip")
		}
	})
}

func TestSlugGeneration(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"2026 Alaska Cruise", "2026-alaska-cruise"},
		{"My  Cool   Trip", "my-cool-trip"},
		{"hello-world", "hello-world"},
		{"UPPER CASE", "upper-case"},
		{"special!@#chars$%^", "special-chars"},
		{"  leading trailing  ", "leading-trailing"},
		{"already-slugged", "already-slugged"},
		{"trip---with---dashes", "trip-with-dashes"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Slugify(tt.name)
			if got != tt.want {
				t.Errorf("Slugify(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}
