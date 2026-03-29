//go:build integration

package notmuch

import (
	"testing"

	"github.com/zachatrocity/voyage/internal/testutil"
)

// These integration tests require a live notmuch database with at least one
// email. They are guarded by the "integration" build tag and will be skipped
// if the notmuch DB is unavailable.
//
// Run with: go test -tags integration ./internal/notmuch/...
//
// Setup requirements:
//   - libnotmuch installed (e.g. `brew install notmuch` or `apt install libnotmuch-dev`)
//   - NOTMUCH_DATABASE env var pointing to a valid mail directory with a notmuch DB
//   - At least one email in the database for the tests to operate on

// firstMessageID returns the message ID of the first email in the database,
// or skips the test if none are available.
func firstMessageID(t *testing.T) string {
	t.Helper()
	results, err := Search("*", "1", SortOldestFirst)
	if err != nil {
		t.Fatalf("Search(*) failed: %v", err)
	}
	if len(results.Results) == 0 {
		t.Skip("no emails in notmuch database")
	}
	return results.Results[0].MessageID
}

func TestIntegration_GetEmailTags(t *testing.T) {
	testutil.SkipIfNoNotmuch(t)

	msgID := firstMessageID(t)

	tags, err := GetEmailTags(msgID)
	if err != nil {
		t.Fatalf("GetEmailTags(%q) error: %v", msgID, err)
	}
	if tags == nil {
		t.Fatalf("GetEmailTags(%q) returned nil (message not found)", msgID)
	}
	// We can't assert specific tags since we don't control the mailbox,
	// but the slice should be non-nil (even if empty).
	t.Logf("tags for %q: %v", msgID, tags)
}

func TestIntegration_GetEmailTags_NotFound(t *testing.T) {
	testutil.SkipIfNoNotmuch(t)

	tags, err := GetEmailTags("nonexistent-message-id-that-does-not-exist@test")
	if err != nil {
		t.Fatalf("GetEmailTags(nonexistent) error: %v", err)
	}
	if tags != nil {
		t.Errorf("expected nil for nonexistent message, got %v", tags)
	}
}

func TestIntegration_TagAndRemoveTag(t *testing.T) {
	testutil.SkipIfNoNotmuch(t)

	msgID := firstMessageID(t)
	testTag := "voyage-test-tag"

	// Add the tag
	result, err := TagEmail(msgID, testTag)
	if err != nil {
		t.Fatalf("TagEmail(%q, %q) error: %v", msgID, testTag, err)
	}
	if result == nil {
		t.Fatalf("TagEmail returned nil (message not found)")
	}

	// Verify tag was added
	found := false
	for _, tag := range result.Tags {
		if tag == testTag {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("tag %q not found after TagEmail; tags: %v", testTag, result.Tags)
	}

	// Remove the tag
	result, err = RemoveTag(msgID, testTag)
	if err != nil {
		t.Fatalf("RemoveTag(%q, %q) error: %v", msgID, testTag, err)
	}
	if result == nil {
		t.Fatalf("RemoveTag returned nil (message not found)")
	}

	// Verify tag was removed
	for _, tag := range result.Tags {
		if tag == testTag {
			t.Errorf("tag %q still present after RemoveTag; tags: %v", testTag, result.Tags)
		}
	}
}

func TestIntegration_RemoveTag_NotFound(t *testing.T) {
	testutil.SkipIfNoNotmuch(t)

	result, err := RemoveTag("nonexistent-message-id-that-does-not-exist@test", "sometag")
	if err != nil {
		t.Fatalf("RemoveTag(nonexistent) error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for nonexistent message, got %+v", result)
	}
}

func TestIntegration_ReplaceTripTag_RoundTrip(t *testing.T) {
	testutil.SkipIfNoNotmuch(t)

	msgID := firstMessageID(t)

	// Tag with trip:foo
	result, err := ReplaceTripTag(msgID, "foo")
	if err != nil {
		t.Fatalf("ReplaceTripTag(%q, foo) error: %v", msgID, err)
	}
	if result == nil {
		t.Fatal("ReplaceTripTag returned nil")
	}

	// Verify trip:foo is present
	hasFoo := false
	for _, tag := range result.Tags {
		if tag == "trip:foo" {
			hasFoo = true
		}
	}
	if !hasFoo {
		t.Errorf("trip:foo not found after ReplaceTripTag; tags: %v", result.Tags)
	}

	// Replace with trip:bar
	result, err = ReplaceTripTag(msgID, "bar")
	if err != nil {
		t.Fatalf("ReplaceTripTag(%q, bar) error: %v", msgID, err)
	}
	if result == nil {
		t.Fatal("ReplaceTripTag returned nil")
	}

	// Verify only trip:bar is present, trip:foo is gone
	hasBar := false
	for _, tag := range result.Tags {
		if tag == "trip:foo" {
			t.Errorf("trip:foo still present after replacing with trip:bar; tags: %v", result.Tags)
		}
		if tag == "trip:bar" {
			hasBar = true
		}
	}
	if !hasBar {
		t.Errorf("trip:bar not found after ReplaceTripTag; tags: %v", result.Tags)
	}

	// Clean up: remove the test tag
	RemoveTag(msgID, "trip:bar")
}

func TestIntegration_ReplaceTripTag_NotFound(t *testing.T) {
	testutil.SkipIfNoNotmuch(t)

	result, err := ReplaceTripTag("nonexistent-message-id-that-does-not-exist@test", "foo")
	if err != nil {
		t.Fatalf("ReplaceTripTag(nonexistent) error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for nonexistent message, got %+v", result)
	}
}

// TestIntegration_AssociateTripEmail_E2E would test the full POST /email/:id/trip/:tripId
// handler end-to-end. This requires both a notmuch database AND a trips store with a
// pre-existing trip.
//
// Setup needed:
//   1. A notmuch database with at least one email (set NOTMUCH_DATABASE env var)
//   2. A trips JSON store with a trip matching the tripId used in the test
//   3. Run with: go test -tags integration ./internal/notmuch/...
//
// The handler-level test is in internal/api/handlers/email_trips_test.go and covers
// the trip-not-found case without needing notmuch. For full E2E testing with a real
// notmuch DB, use the ReplaceTripTag round-trip test above which exercises the same
// core logic.
