package trips

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type Trip struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	DateRange string    `json:"date_range"`
	CreatedAt time.Time `json:"created_at"`
}

type Store interface {
	Create(t *Trip) error
	List() ([]*Trip, error)
	Get(id string) (*Trip, error)
	Delete(id string) error
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

var (
	slugNonAlnum = regexp.MustCompile(`[^a-z0-9-]+`)
	slugMultiDash = regexp.MustCompile(`-{2,}`)
)

func Slugify(name string) string {
	s := strings.ToLower(name)
	s = slugNonAlnum.ReplaceAllString(s, "-")
	s = slugMultiDash.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

func (s *Service) CreateTrip(name, dateRange string) (*Trip, error) {
	id := Slugify(name)
	if id == "" {
		return nil, fmt.Errorf("invalid trip name: %q", name)
	}
	t := &Trip{
		ID:        id,
		Name:      name,
		DateRange: dateRange,
		CreatedAt: time.Now(),
	}
	if err := s.store.Create(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) ListTrips() ([]*Trip, error) {
	return s.store.List()
}

func (s *Service) GetTrip(id string) (*Trip, error) {
	return s.store.Get(id)
}

func (s *Service) DeleteTrip(id string) error {
	return s.store.Delete(id)
}
