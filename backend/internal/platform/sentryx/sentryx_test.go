package sentryx

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDisabledByDefault(t *testing.T) {
	if Enabled() {
		t.Fatalf("expected Sentry disabled without Init")
	}
	Capture(context.Background(), errors.New("boom"))
	CaptureMessage(context.Background(), "hello")
	Flush(time.Second)
}

func TestInitEmptyDSNStaysDisabled(t *testing.T) {
	if err := Init("", "test"); err != nil {
		t.Fatalf("empty DSN must not error: %v", err)
	}
	if Enabled() {
		t.Fatalf("expected Sentry disabled with empty DSN")
	}
}

func TestInitInvalidDSNFails(t *testing.T) {
	if err := Init("://not-a-dsn", "test"); err == nil {
		t.Fatalf("expected error for malformed DSN")
	}
}
