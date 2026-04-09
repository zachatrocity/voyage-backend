package classifier

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// Category represents the type of travel email.
type Category string

const (
	CategoryFlight    Category = "flight"
	CategoryHotel     Category = "hotel"
	CategoryCarRental Category = "car_rental"
	CategoryCruise    Category = "cruise"
	CategoryActivity  Category = "activity"
	CategoryOther     Category = "other"
)

// CategoryRule holds the matching rules for a single category.
type CategoryRule struct {
	Domains         []string `yaml:"domains" json:"domains"`
	SubjectKeywords []string `yaml:"subject_keywords" json:"subject_keywords"`
}

// ClassifiersConfig is the top-level classifier configuration structure.
type ClassifiersConfig struct {
	Categories map[string]CategoryRule `yaml:"categories" json:"categories"`
}

// DefaultConfig is the built-in fallback when no external file is found.
// Exported so handlers can reference it directly.
var DefaultConfig = ClassifiersConfig{
	Categories: map[string]CategoryRule{
		"flight": {
			Domains:         []string{"delta.com", "united.com", "aa.com", "southwest.com", "alaskaair.com", "spirit.com", "jetblue.com", "frontier.com"},
			SubjectKeywords: []string{"flight", "eticket", "e-ticket", "boarding", "confirmation", "departing", "arrives", "itinerary"},
		},
		"hotel": {
			Domains:         []string{"marriott.com", "hilton.com", "hyatt.com", "airbnb.com", "vrbo.com", "booking.com", "expedia.com", "ihg.com"},
			SubjectKeywords: []string{"hotel", "check-in", "checkout", "reservation", "stay", "resort", "lodging"},
		},
		"cruise": {
			Domains:         []string{"carnival.com", "royalcaribbean.com", "ncl.com", "princess.com", "celebritycruises.com"},
			SubjectKeywords: []string{"cruise", "embarkation", "port", "stateroom", "sailing", "voyage", "ship"},
		},
		"car_rental": {
			Domains:         []string{"hertz.com", "enterprise.com", "avis.com", "budget.com", "nationalcar.com", "alamo.com"},
			SubjectKeywords: []string{"rental", "pick-up", "vehicle", "car reservation", "pickup"},
		},
		"activity": {
			Domains:         []string{"disney.com", "universalorlando.com", "ticketmaster.com", "viator.com", "getyourguide.com"},
			SubjectKeywords: []string{"ticket", "admission", "park", "tour", "experience", "booking"},
		},
	},
}

// Package-level state for injectable config path and in-memory cache.
var (
	mu           sync.RWMutex
	configPath   string
	cachedConfig *ClassifiersConfig
)

// SetConfigPath sets the classifiers YAML file path and loads it into memory.
// Call this at application startup to inject the config path from main.go.
// Falls back to DefaultConfig on read/parse error (returns the error for logging).
func SetConfigPath(path string) error {
	mu.Lock()
	defer mu.Unlock()
	configPath = path
	cfg, err := loadConfigFromPath(path)
	if err != nil {
		// Cache default so DetectCategory still works
		dflt := DefaultConfig
		cachedConfig = &dflt
		return err
	}
	cachedConfig = &cfg
	return nil
}

// GetConfigPath returns the currently configured classifiers file path.
func GetConfigPath() string {
	mu.RLock()
	defer mu.RUnlock()
	return configPath
}

// GetConfig returns the current in-memory classifiers config.
// If SetConfigPath has been called, returns the cached config.
// Otherwise falls back to VOYAGE_CLASSIFIERS_PATH env var or DefaultConfig.
func GetConfig() ClassifiersConfig {
	return getActiveConfig()
}

// ReloadConfig reloads the config from the current configPath.
// Returns an error if no path has been set or if loading fails.
func ReloadConfig() error {
	mu.Lock()
	defer mu.Unlock()
	if configPath == "" {
		return fmt.Errorf("no config path set; call SetConfigPath first")
	}
	cfg, err := loadConfigFromPath(configPath)
	if err != nil {
		return err
	}
	cachedConfig = &cfg
	return nil
}

// WriteConfig marshals cfg to YAML and writes it to path.
func WriteConfig(path string, cfg ClassifiersConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	return nil
}

// WriteDefaultConfig writes the built-in DefaultConfig as YAML to path.
func WriteDefaultConfig(path string) error {
	return WriteConfig(path, DefaultConfig)
}

// getActiveConfig returns the effective config, respecting injection vs legacy env var.
func getActiveConfig() ClassifiersConfig {
	mu.RLock()
	path := configPath
	cached := cachedConfig
	mu.RUnlock()

	// If SetConfigPath was called, use the cached config.
	if path != "" {
		if cached != nil {
			return *cached
		}
		// Path set but cache empty (race or reset); reload.
		cfg, _ := loadConfigFromPath(path)
		return cfg
	}

	// Legacy fallback: read VOYAGE_CLASSIFIERS_PATH env var each call.
	// This preserves backward compatibility for tests that use t.Setenv.
	return loadConfigFromEnvOrDefault()
}

// loadConfigFromEnvOrDefault reads from VOYAGE_CLASSIFIERS_PATH or returns DefaultConfig.
func loadConfigFromEnvOrDefault() ClassifiersConfig {
	envPath := os.Getenv("VOYAGE_CLASSIFIERS_PATH")
	if envPath == "" {
		return DefaultConfig
	}
	cfg, err := loadConfigFromPath(envPath)
	if err != nil {
		return DefaultConfig
	}
	return cfg
}

// loadConfigFromPath reads and parses a YAML classifiers file.
func loadConfigFromPath(path string) (ClassifiersConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultConfig, fmt.Errorf("read classifiers file %q: %w", path, err)
	}
	var cfg ClassifiersConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig, fmt.Errorf("parse classifiers file %q: %w", path, err)
	}
	return cfg, nil
}

// extractDomain returns the domain portion of an email address (lowercased).
func extractDomain(from string) string {
	// Handle "Name <email@domain>" format
	if idx := strings.LastIndex(from, "<"); idx != -1 {
		from = from[idx+1:]
		from = strings.TrimRight(from, ">")
	}
	if at := strings.LastIndex(from, "@"); at != -1 {
		return strings.ToLower(from[at+1:])
	}
	return strings.ToLower(from)
}

// categoryPriority defines the order in which categories are evaluated for
// subject-keyword matching. More specific categories come first so that e.g.
// "rental" beats the generic "confirmation" (flight) keyword.
var categoryPriority = []string{"cruise", "car_rental", "hotel", "activity", "flight"}

// DetectCategory classifies an email based on the sender domain and subject line.
// Domain matching takes priority; subject keyword matching is the fallback.
func DetectCategory(subject, from string) Category {
	cfg := getActiveConfig()
	domain := extractDomain(from)

	// First pass: match by domain (iterate in priority order for determinism)
	for _, cat := range categoryPriority {
		rule, ok := cfg.Categories[cat]
		if !ok {
			continue
		}
		for _, d := range rule.Domains {
			if strings.ToLower(d) == domain {
				return Category(cat)
			}
		}
	}

	// Second pass: match by subject keywords in priority order
	lowerSubject := strings.ToLower(subject)
	for _, cat := range categoryPriority {
		rule, ok := cfg.Categories[cat]
		if !ok {
			continue
		}
		for _, kw := range rule.SubjectKeywords {
			if strings.Contains(lowerSubject, strings.ToLower(kw)) {
				return Category(cat)
			}
		}
	}

	return CategoryOther
}
