package notmuch

import (
	"testing"
)

func TestFilterTripTags_Empty(t *testing.T) {
	result := FilterTripTags([]string{})
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %v", result)
	}
}

func TestFilterTripTags_NoTripTags(t *testing.T) {
	tags := []string{"inbox", "unread", "travel", "important"}
	result := FilterTripTags(tags)
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %v", result)
	}
}

func TestFilterTripTags_SingleTripTag(t *testing.T) {
	tags := []string{"inbox", "trip:2026-cruise", "travel"}
	result := FilterTripTags(tags)
	if len(result) != 1 || result[0] != "trip:2026-cruise" {
		t.Errorf("expected [trip:2026-cruise], got %v", result)
	}
}

func TestFilterTripTags_MultipleTripTags(t *testing.T) {
	tags := []string{"trip:old-trip", "inbox", "trip:2026-cruise", "travel"}
	result := FilterTripTags(tags)
	if len(result) != 2 {
		t.Fatalf("expected 2 trip tags, got %d: %v", len(result), result)
	}
	if result[0] != "trip:old-trip" {
		t.Errorf("result[0] = %q; want trip:old-trip", result[0])
	}
	if result[1] != "trip:2026-cruise" {
		t.Errorf("result[1] = %q; want trip:2026-cruise", result[1])
	}
}

func TestFilterTripTags_TripPrefixOnly(t *testing.T) {
	// "trip:" with no value and "tripper" should not match
	tags := []string{"trip:", "tripper", "trip:valid"}
	result := FilterTripTags(tags)
	// "trip:" has the prefix so it matches; "tripper" does not
	if len(result) != 2 {
		t.Fatalf("expected 2 trip tags, got %d: %v", len(result), result)
	}
	if result[0] != "trip:" {
		t.Errorf("result[0] = %q; want trip:", result[0])
	}
	if result[1] != "trip:valid" {
		t.Errorf("result[1] = %q; want trip:valid", result[1])
	}
}

func TestFilterTripTags_NilInput(t *testing.T) {
	result := FilterTripTags(nil)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}
