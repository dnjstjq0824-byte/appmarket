// Package sanitize is the single XSS defense for user-controlled text.
// Every write path that persists or echoes free-form input must pass that
// input through Strict (plain text) or UGC (limited rich text) before
// storage. JSON encoding and frontend auto-escaping are rendering concerns;
// this package guarantees stored values contain no executable markup.
package sanitize

import (
	"sync"

	"github.com/microcosm-cc/bluemonday"
)

var (
	strictOnce sync.Once
	strict     *bluemonday.Policy
	ugcOnce    sync.Once
	ugc        *bluemonday.Policy
)

// Strict strips all HTML/XML markup and returns plain text.
func Strict(input string) string {
	strictOnce.Do(func() { strict = bluemonday.StrictPolicy() })
	return strict.Sanitize(input)
}

// UGC allows the bluemonday user-generated-content subset (links, emphasis,
// lists) and strips scripts, styles, event handlers, and unsafe URLs.
func UGC(input string) string {
	ugcOnce.Do(func() { ugc = bluemonday.UGCPolicy() })
	return ugc.Sanitize(input)
}
