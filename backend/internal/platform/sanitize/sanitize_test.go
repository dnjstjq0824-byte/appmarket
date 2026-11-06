package sanitize

import (
	"strings"
	"testing"
)

func TestStrictStripsExecutableMarkup(t *testing.T) {
	out := Strict(`Hello <script>alert(1)</script><img src=x onerror=alert(2)> <b>bold</b>`)
	if strings.Contains(out, "<script") || strings.Contains(out, "onerror") || strings.Contains(out, "<img") {
		t.Fatalf("executable markup survived strict sanitize: %q", out)
	}
	if !strings.Contains(out, "Hello") || !strings.Contains(out, "bold") {
		t.Fatalf("plain text lost by strict sanitize: %q", out)
	}
}

func TestStrictKeepsPlainText(t *testing.T) {
	in := "5 > 3 & 2 < 4 — plain review text"
	out := Strict(in)
	if !strings.Contains(out, "plain review text") {
		t.Fatalf("plain text altered: %q", out)
	}
}

func TestUGCAllowsSafeSubsetOnly(t *testing.T) {
	out := UGC(`<a href="https://example.com" onclick="evil()">link</a><script>evil()</script>`)
	if strings.Contains(out, "onclick") || strings.Contains(out, "<script") {
		t.Fatalf("unsafe markup survived UGC sanitize: %q", out)
	}
	if !strings.Contains(out, "https://example.com") {
		t.Fatalf("safe link lost by UGC sanitize: %q", out)
	}
}
