package classifier

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectCategory_DomainMatch(t *testing.T) {
	tests := []struct {
		name     string
		subject  string
		from     string
		expected Category
	}{
		{"flight by domain", "Your itinerary", "noreply@delta.com", CategoryFlight},
		{"hotel by domain", "Reservation confirmed", "booking@marriott.com", CategoryHotel},
		{"cruise by domain", "Welcome aboard", "info@carnival.com", CategoryCruise},
		{"car rental by domain", "Your rental", "confirm@hertz.com", CategoryCarRental},
		{"activity by domain", "Your tickets", "tickets@disney.com", CategoryActivity},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectCategory(tt.subject, tt.from)
			if got != tt.expected {
				t.Errorf("DetectCategory(%q, %q) = %q, want %q", tt.subject, tt.from, got, tt.expected)
			}
		})
	}
}

func TestDetectCategory_SubjectKeyword(t *testing.T) {
	tests := []struct {
		name     string
		subject  string
		from     string
		expected Category
	}{
		{"flight by subject", "Your Flight Confirmation", "noreply@unknown.com", CategoryFlight},
		{"hotel by subject", "Hotel Reservation Details", "noreply@unknown.com", CategoryHotel},
		{"cruise by subject", "Cruise Embarkation Info", "noreply@unknown.com", CategoryCruise},
		{"car rental by subject", "Your Rental Confirmation", "noreply@unknown.com", CategoryCarRental},
		{"activity by subject", "Your Admission Pass", "noreply@unknown.com", CategoryActivity},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectCategory(tt.subject, tt.from)
			if got != tt.expected {
				t.Errorf("DetectCategory(%q, %q) = %q, want %q", tt.subject, tt.from, got, tt.expected)
			}
		})
	}
}

func TestDetectCategory_FallbackOther(t *testing.T) {
	got := DetectCategory("Hello world", "someone@example.com")
	if got != CategoryOther {
		t.Errorf("expected CategoryOther, got %q", got)
	}
}

func TestDetectCategory_DomainPriority(t *testing.T) {
	// Domain should win even if subject contains keywords for another category
	got := DetectCategory("Your Hotel Reservation", "noreply@delta.com")
	if got != CategoryFlight {
		t.Errorf("expected domain match (flight), got %q", got)
	}
}

func TestDetectCategory_AngleBracketFrom(t *testing.T) {
	got := DetectCategory("Booking confirmed", "Delta Airlines <noreply@delta.com>")
	if got != CategoryFlight {
		t.Errorf("expected flight, got %q", got)
	}
}

func TestDetectCategory_CaseInsensitive(t *testing.T) {
	got := DetectCategory("YOUR FLIGHT DEPARTING SOON", "unknown@example.com")
	if got != CategoryFlight {
		t.Errorf("expected flight from uppercase subject, got %q", got)
	}
}

func TestDetectCategory_CustomConfigFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "classifiers.yaml")
	content := `categories:
  flight:
    domains:
      - customair.com
    subject_keywords:
      - custom flight
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	SetConfigPath("")
	t.Setenv("VOYAGE_CLASSIFIERS_PATH", cfgPath)

	got := DetectCategory("anything", "info@customair.com")
	if got != CategoryFlight {
		t.Errorf("expected flight from custom config, got %q", got)
	}

	// Default domains should NOT match when custom config is loaded
	got = DetectCategory("anything", "info@delta.com")
	if got == CategoryFlight {
		t.Error("delta.com should not match in custom config")
	}
}
