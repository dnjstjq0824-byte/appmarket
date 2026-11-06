// Package sentryx wraps the Sentry error-tracking SDK with an explicit
// enabled/disabled switch. An empty DSN disables reporting entirely; every
// other function is a safe no-op until Init succeeds.
package sentryx

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
)

var (
	mu      sync.RWMutex
	enabled bool
)

// Init configures Sentry. Empty dsn leaves reporting disabled.
func Init(dsn, environment string) error {
	if dsn == "" {
		return nil
	}
	if environment == "" {
		environment = "development"
	}
	err := sentry.Init(sentry.ClientOptions{
		Dsn:         dsn,
		Environment: environment,
	})
	if err != nil {
		return err
	}
	mu.Lock()
	enabled = true
	mu.Unlock()
	return nil
}

// Enabled reports whether error reporting is active.
func Enabled() bool {
	mu.RLock()
	defer mu.RUnlock()
	return enabled
}

// Capture sends err to Sentry with the request trace id when enabled.
func Capture(ctx context.Context, err error) {
	if err == nil {
		return
	}
	mu.RLock()
	on := enabled
	mu.RUnlock()
	if !on {
		return
	}
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub()
	}
	hub.CaptureException(err)
}

// CaptureMessage sends a plain message to Sentry when enabled.
func CaptureMessage(ctx context.Context, msg string) {
	if msg == "" {
		return
	}
	Capture(ctx, errors.New(msg))
}

// Flush waits for buffered events to ship.
func Flush(timeout time.Duration) {
	mu.RLock()
	on := enabled
	mu.RUnlock()
	if !on {
		return
	}
	sentry.Flush(timeout)
}
