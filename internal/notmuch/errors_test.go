package notmuch

import (
	"errors"
	"testing"
)

func TestAsOperationError(t *testing.T) {
	base := errors.New("boom")
	err := &OperationError{
		Operation:    "replace_trip_tag_open",
		Code:         "notmuch_db_open_xapian_exception",
		Retryable:    true,
		DatabasePath: "/mail",
		Mode:         "read_write",
		Cause:        base,
	}

	got, ok := AsOperationError(err)
	if !ok {
		t.Fatal("expected AsOperationError to match")
	}
	if got.Code != "notmuch_db_open_xapian_exception" {
		t.Fatalf("unexpected code: %s", got.Code)
	}
	if !got.Retryable {
		t.Fatal("expected retryable=true")
	}
}

func TestOperationErrorDetails(t *testing.T) {
	err := &OperationError{
		Operation:    "tag_email_open",
		Code:         "notmuch_db_open_failed",
		Retryable:    false,
		DatabasePath: "/mail",
		Mode:         "read_write",
		Cause:        errors.New("failed to open notmuch database"),
	}

	details := err.Details()
	if details["operation"] != "tag_email_open" {
		t.Fatalf("unexpected operation detail: %v", details["operation"])
	}
	if details["mode"] != "read_write" {
		t.Fatalf("unexpected mode detail: %v", details["mode"])
	}
}
