package notmuch

import (
	"testing"
	"time"

	libnotmuch "github.com/zachatrocity/voyage/notmuch"
)

func TestDBOpenRetryConfig_ReadOnly(t *testing.T) {
	attempts, delay := dbOpenRetryConfig(libnotmuch.DATABASE_MODE_READ_ONLY)
	if attempts != readOnlyDBOpenMaxAttempts {
		t.Fatalf("unexpected read-only attempts: got %d want %d", attempts, readOnlyDBOpenMaxAttempts)
	}
	if delay != readOnlyDBOpenRetryDelay {
		t.Fatalf("unexpected read-only delay: got %s want %s", delay, readOnlyDBOpenRetryDelay)
	}
}

func TestDBOpenRetryConfig_ReadWrite(t *testing.T) {
	attempts, delay := dbOpenRetryConfig(libnotmuch.DATABASE_MODE_READ_WRITE)
	if attempts != readWriteDBOpenMaxAttempts {
		t.Fatalf("unexpected read-write attempts: got %d want %d", attempts, readWriteDBOpenMaxAttempts)
	}
	if delay != readWriteDBOpenRetryDelay {
		t.Fatalf("unexpected read-write delay: got %s want %s", delay, readWriteDBOpenRetryDelay)
	}
	if attempts <= readOnlyDBOpenMaxAttempts {
		t.Fatalf("expected read-write attempts > read-only attempts, got %d <= %d", attempts, readOnlyDBOpenMaxAttempts)
	}
	if delay < 150*time.Millisecond {
		t.Fatalf("expected read-write retry delay to be at least 150ms, got %s", delay)
	}
}
