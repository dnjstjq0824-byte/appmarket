package domain

import "testing"

func TestListCursorRoundTrip(t *testing.T) {
	enc := EncodeListCursor(4.825, 125000, "app-omni-01")
	if enc == "" {
		t.Fatalf("expected non-empty cursor")
	}
	r, n, id, ok := DecodeListCursor(enc)
	if !ok {
		t.Fatalf("failed to decode cursor %q", enc)
	}
	if r != 4.825 || n != 125000 || id != "app-omni-01" {
		t.Fatalf("round trip mismatch: got (%v, %d, %q)", r, n, id)
	}
}

func TestListCursorMalformedIsFirstPage(t *testing.T) {
	for _, bad := range []string{"", "not-base64!!!", "e30=", "bm9uLWpzb24"} {
		if _, _, _, ok := DecodeListCursor(bad); ok {
			t.Fatalf("expected ok=false for %q", bad)
		}
	}
}

func TestAfterCursorOrdering(t *testing.T) {
	if !AfterCursor(4.8, 10, "b", 4.9, 10, "a") {
		t.Fatalf("lower rating must sort after")
	}
	if AfterCursor(4.9, 10, "b", 4.8, 10, "a") {
		t.Fatalf("higher rating must not sort after")
	}
	if !AfterCursor(4.9, 5, "b", 4.9, 10, "a") {
		t.Fatalf("fewer installs must sort after on rating tie")
	}
	if !AfterCursor(4.9, 10, "b", 4.9, 10, "a") {
		t.Fatalf("greater id must sort after on full tie")
	}
	if AfterCursor(4.9, 10, "a", 4.9, 10, "a") {
		t.Fatalf("identical row must not sort after itself")
	}
}
