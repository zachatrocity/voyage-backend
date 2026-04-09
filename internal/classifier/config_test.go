package classifier

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAndReloadConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "classifiers.yaml")
	SetConfigPath(path)

	cfg := ClassifiersConfig{
		Categories: map[string]CategoryRule{
			"flight": {
				Domains:         []string{"exampleair.com"},
				SubjectKeywords: []string{"custom boarding"},
			},
		},
	}

	if err := WriteConfig(cfg); err != nil {
		t.Fatalf("WriteConfig error: %v", err)
	}

	reloaded := ReloadConfig()
	if got := len(reloaded.Categories); got != 1 {
		t.Fatalf("expected 1 category, got %d", got)
	}
	if got := reloaded.Categories["flight"].Domains[0]; got != "exampleair.com" {
		t.Fatalf("unexpected domain: %s", got)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected config file to exist: %v", err)
	}
}

func TestWriteConfigRejectsInvalidAndPreservesPrevious(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "classifiers.yaml")
	SetConfigPath(path)

	valid := ClassifiersConfig{
		Categories: map[string]CategoryRule{
			"flight": {Domains: []string{"exampleair.com"}},
		},
	}
	if err := WriteConfig(valid); err != nil {
		t.Fatalf("WriteConfig valid error: %v", err)
	}

	invalid := ClassifiersConfig{
		Categories: map[string]CategoryRule{
			"flight": {Domains: []string{""}},
		},
	}
	if err := WriteConfig(invalid); err == nil {
		t.Fatal("expected validation error for invalid config")
	}

	cfg := GetConfig()
	if got := cfg.Categories["flight"].Domains[0]; got != "exampleair.com" {
		t.Fatalf("expected previous valid config to remain, got %q", got)
	}
}

func TestWriteDefaultConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "classifiers.yaml")
	SetConfigPath(path)

	cfg, err := WriteDefaultConfig()
	if err != nil {
		t.Fatalf("WriteDefaultConfig error: %v", err)
	}

	if len(cfg.Categories) == 0 {
		t.Fatal("expected default categories")
	}
	if _, ok := cfg.Categories["flight"]; !ok {
		t.Fatal("expected flight category in defaults")
	}
}
