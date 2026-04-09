package classifier

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

// CategoryRule defines domain and subject matching rules for one category.
type CategoryRule struct {
	Domains         []string `json:"domains" yaml:"domains"`
	SubjectKeywords []string `json:"subject_keywords" yaml:"subject_keywords"`
}

// ClassifiersConfig defines all classifier categories and their matching rules.
type ClassifiersConfig struct {
	Categories map[string]CategoryRule `json:"categories" yaml:"categories"`
}

var categoryPriority = []string{"cruise", "car_rental", "hotel", "activity", "flight"}

var defaultConfig = ClassifiersConfig{
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

var (
	configPathMu sync.RWMutex
	configPath   string

	cachedConfigMu sync.RWMutex
	cachedConfig   *ClassifiersConfig
)

// SetConfigPath sets the yaml path used for classifier configuration.
// Empty path means defaults only.
func SetConfigPath(path string) {
	configPathMu.Lock()
	configPath = strings.TrimSpace(path)
	configPathMu.Unlock()

	cachedConfigMu.Lock()
	cachedConfig = nil
	cachedConfigMu.Unlock()
}

func resolveConfigPath() string {
	configPathMu.RLock()
	defer configPathMu.RUnlock()
	if configPath != "" {
		return configPath
	}
	if v := strings.TrimSpace(os.Getenv("VOYAGE_CLASSIFIERS")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("VOYAGE_CLASSIFIERS_PATH")); v != "" {
		return v
	}
	return ""
}

func cloneConfig(cfg ClassifiersConfig) ClassifiersConfig {
	out := ClassifiersConfig{Categories: make(map[string]CategoryRule, len(cfg.Categories))}
	for k, v := range cfg.Categories {
		rule := CategoryRule{
			Domains:         append([]string(nil), v.Domains...),
			SubjectKeywords: append([]string(nil), v.SubjectKeywords...),
		}
		out.Categories[k] = rule
	}
	return out
}

func normalizeConfig(cfg ClassifiersConfig) ClassifiersConfig {
	out := ClassifiersConfig{Categories: make(map[string]CategoryRule, len(cfg.Categories))}
	for category, rule := range cfg.Categories {
		name := strings.TrimSpace(strings.ToLower(category))
		if name == "" {
			continue
		}

		normRule := CategoryRule{}
		seenDomains := map[string]struct{}{}
		for _, d := range rule.Domains {
			domain := strings.TrimSpace(strings.ToLower(d))
			if domain == "" {
				continue
			}
			if _, ok := seenDomains[domain]; ok {
				continue
			}
			seenDomains[domain] = struct{}{}
			normRule.Domains = append(normRule.Domains, domain)
		}

		seenKeywords := map[string]struct{}{}
		for _, kw := range rule.SubjectKeywords {
			keyword := strings.TrimSpace(strings.ToLower(kw))
			if keyword == "" {
				continue
			}
			if _, ok := seenKeywords[keyword]; ok {
				continue
			}
			seenKeywords[keyword] = struct{}{}
			normRule.SubjectKeywords = append(normRule.SubjectKeywords, keyword)
		}

		out.Categories[name] = normRule
	}
	return out
}

// ValidateConfig validates classifiers config and returns a descriptive error.
func ValidateConfig(cfg ClassifiersConfig) error {
	if len(cfg.Categories) == 0 {
		return errors.New("categories must contain at least one category")
	}

	for rawCategory, rule := range cfg.Categories {
		category := strings.TrimSpace(rawCategory)
		if category == "" {
			return errors.New("category name cannot be empty")
		}
		if len(rule.Domains) == 0 && len(rule.SubjectKeywords) == 0 {
			return fmt.Errorf("category %q must include at least one domain or subject keyword", category)
		}

		for _, d := range rule.Domains {
			if strings.TrimSpace(d) == "" {
				return fmt.Errorf("category %q contains an empty domain", category)
			}
		}
		for _, kw := range rule.SubjectKeywords {
			if strings.TrimSpace(kw) == "" {
				return fmt.Errorf("category %q contains an empty subject keyword", category)
			}
		}
	}

	return nil
}

// GetDefaultConfig returns built-in defaults.
func GetDefaultConfig() ClassifiersConfig {
	return cloneConfig(defaultConfig)
}

func loadConfigFromDisk() ClassifiersConfig {
	path := resolveConfigPath()
	if path == "" {
		return cloneConfig(defaultConfig)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return cloneConfig(defaultConfig)
	}

	var cfg ClassifiersConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cloneConfig(defaultConfig)
	}

	cfg = normalizeConfig(cfg)
	if err := ValidateConfig(cfg); err != nil {
		return cloneConfig(defaultConfig)
	}

	return cfg
}

// ReloadConfig reloads config from disk (or defaults) into cache.
func ReloadConfig() ClassifiersConfig {
	cfg := loadConfigFromDisk()
	cachedConfigMu.Lock()
	cachedConfig = &cfg
	cachedConfigMu.Unlock()
	return cloneConfig(cfg)
}

// GetConfig returns currently active config from cache, loading if needed.
func GetConfig() ClassifiersConfig {
	cachedConfigMu.RLock()
	if cachedConfig != nil {
		cfg := cloneConfig(*cachedConfig)
		cachedConfigMu.RUnlock()
		return cfg
	}
	cachedConfigMu.RUnlock()

	return ReloadConfig()
}

// WriteConfig validates and persists config to disk, preserving prior config on failure.
func WriteConfig(cfg ClassifiersConfig) error {
	cfg = normalizeConfig(cfg)
	if err := ValidateConfig(cfg); err != nil {
		return err
	}

	path := resolveConfigPath()
	if path == "" {
		return errors.New("classifier config path is not configured")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create classifiers dir: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal classifiers config: %w", err)
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("write temp classifiers config: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("persist classifiers config: %w", err)
	}

	cachedConfigMu.Lock()
	copyCfg := cloneConfig(cfg)
	cachedConfig = &copyCfg
	cachedConfigMu.Unlock()
	return nil
}

// WriteDefaultConfig writes default config to disk and updates cache.
func WriteDefaultConfig() (ClassifiersConfig, error) {
	cfg := GetDefaultConfig()
	if err := WriteConfig(cfg); err != nil {
		return ClassifiersConfig{}, err
	}
	return cfg, nil
}

// extractDomain returns the domain portion of an email address (lowercased).
func extractDomain(from string) string {
	if idx := strings.LastIndex(from, "<"); idx != -1 {
		from = from[idx+1:]
		from = strings.TrimRight(from, ">")
	}
	if at := strings.LastIndex(from, "@"); at != -1 {
		return strings.ToLower(from[at+1:])
	}
	return strings.ToLower(from)
}

// DetectCategory classifies an email based on the sender domain and subject line.
// Domain matching takes priority; subject keyword matching is the fallback.
func DetectCategory(subject, from string) Category {
	cfg := GetConfig()
	domain := extractDomain(from)

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
