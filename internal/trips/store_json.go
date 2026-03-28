package trips

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type JSONStore struct {
	path string
}

func NewJSONStore(path string) (*JSONStore, error) {
	if path == "" {
		path = os.Getenv("VOYAGE_TRIPS_PATH")
	}
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("determining home directory: %w", err)
		}
		path = filepath.Join(home, ".config", "voyage", "trips.json")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("creating directory %s: %w", dir, err)
	}
	return &JSONStore{path: path}, nil
}

func (s *JSONStore) readAll() ([]*Trip, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	var trips []*Trip
	if err := json.Unmarshal(data, &trips); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", s.path, err)
	}
	return trips, nil
}

func (s *JSONStore) writeAll(trips []*Trip) error {
	data, err := json.MarshalIndent(trips, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

func (s *JSONStore) Create(t *Trip) error {
	trips, err := s.readAll()
	if err != nil {
		return err
	}
	for _, existing := range trips {
		if existing.ID == t.ID {
			return fmt.Errorf("trip %q already exists", t.ID)
		}
	}
	trips = append(trips, t)
	return s.writeAll(trips)
}

func (s *JSONStore) List() ([]*Trip, error) {
	trips, err := s.readAll()
	if err != nil {
		return nil, err
	}
	if trips == nil {
		return []*Trip{}, nil
	}
	return trips, nil
}

func (s *JSONStore) Get(id string) (*Trip, error) {
	trips, err := s.readAll()
	if err != nil {
		return nil, err
	}
	for _, t := range trips {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, fmt.Errorf("trip %q not found", id)
}

func (s *JSONStore) Delete(id string) error {
	trips, err := s.readAll()
	if err != nil {
		return err
	}
	for i, t := range trips {
		if t.ID == id {
			trips = append(trips[:i], trips[i+1:]...)
			return s.writeAll(trips)
		}
	}
	return fmt.Errorf("trip %q not found", id)
}
