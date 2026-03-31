package classifier

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Category represents the type of travel email.
type Category string

const (
	CategoryFlight   Category = "flight"
	CategoryHotel    Category = "hotel"
	CategoryCarRental Category = "car_rental"
	CategoryCruise   Category = "cruise"
	CategoryActivity Category = "activity"
	CategoryOther    Category = "other"
)

type categoryRule struct {
	Domains         []string `yaml:"domains"`
	SubjectKeywords []string `yaml:"subject_keywords"`
}

type classifiersConfig struct {
	Categories map[string]categoryRule `yaml:"categories"`
}

// defaultConfig is the built-in fallback when no external file is found.
var defaultConfig = classifiersConfig{
	Categories: map[string]categoryRule{
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

func loadConfig() classifiersConfig {
	path := os.Getenv("VOYAGE_CLASSIFIERS_PATH")
	if path == "" {
		return defaultConfig
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return defaultConfig
	}
	var cfg classifiersConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return defaultConfig
	}
	return cfg
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
	cfg := loadConfig()
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
