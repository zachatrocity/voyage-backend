package testutil

import (
	"testing"

	voyagenotmuch "github.com/zachatrocity/voyage/internal/notmuch"
)

// SkipIfNoNotmuch skips the test if the notmuch database is not available.
// Use this in integration tests that require a live notmuch DB.
func SkipIfNoNotmuch(t *testing.T) {
	t.Helper()
	if err := voyagenotmuch.CheckDatabaseConnection(); err != nil {
		t.Skip("notmuch DB not available: " + err.Error())
	}
}
