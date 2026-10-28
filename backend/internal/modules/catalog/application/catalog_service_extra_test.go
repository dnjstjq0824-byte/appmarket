package application_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/domain"
)

func TestCatalogSignatureGroupSeparation(t *testing.T) {
	svc, repo, _ := setupCatalog(t)
	locs := []domain.Localization{{ID: "l1", AppID: "app-sig", Locale: "en-US", Title: "Sig App"}}
	dists := []domain.Distribution{
		{ID: "d1", AppID: "app-sig", SourceType: domain.SourceBNHosted, DistributionMode: domain.ModeDirectAPK, PackageName: "com.example.sig", SigningCertSHA256: "cert-AAA"},
		{ID: "d2", AppID: "app-sig", SourceType: domain.SourceFDroidRepo, DistributionMode: domain.ModeFederatedRepoBin, PackageName: "com.example.sig", SigningCertSHA256: "cert-BBB"},
	}
	seedApp(repo, "sig", "en-US", locs, dists)

	detail, err := svc.GetAppDetail(context.Background(), "sig", "en-US")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(detail.Distributions) != 2 {
		t.Fatalf("expected 2 distributions, got %d", len(detail.Distributions))
	}
	d0 := detail.Distributions[0]
	d1 := detail.Distributions[1]
	if d0.IsSignatureCompatible(d1) {
		t.Fatalf("distributions with different certs must not be signature-compatible")
	}
	if d0.SignatureCompatibilityGroup == d1.SignatureCompatibilityGroup {
		t.Fatalf("signature compatibility groups must differ: %s vs %s", d0.SignatureCompatibilityGroup, d1.SignatureCompatibilityGroup)
	}
}

func TestCatalogXSSSanitization(t *testing.T) {
	svc, repo, _ := setupCatalog(t)
	locs := []domain.Localization{
		{ID: "l1", AppID: "app-xss", Locale: "en-US", Title: "XSS App", FullDescription: `<p>Hello</p><script>alert('xss')</script><img src=x onerror=alert(1)>`},
	}
	dists := []domain.Distribution{{ID: "d1", AppID: "app-xss", SourceType: domain.SourceBNHosted, SigningCertSHA256: "cert"}}
	seedApp(repo, "xss", "en-US", locs, dists)

	detail, err := svc.GetAppDetail(context.Background(), "xss", "en-US")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if detail.Localization == nil {
		t.Fatal("expected localization")
	}
	if strings.Contains(detail.Localization.FullDescription, "<script") {
		t.Fatalf("script tag not sanitized: %s", detail.Localization.FullDescription)
	}
	if strings.Contains(detail.Localization.FullDescription, "onerror") {
		t.Fatalf("onerror not sanitized: %s", detail.Localization.FullDescription)
	}
	if !strings.Contains(detail.Localization.FullDescription, "Hello") {
		t.Fatalf("content lost after sanitization: %s", detail.Localization.FullDescription)
	}
}

func TestCatalogCacheHit(t *testing.T) {
	svc, repo, cache := setupCatalog(t)
	locs := []domain.Localization{{ID: "l1", AppID: "app-cache", Locale: "en-US", Title: "Cache App"}}
	dists := []domain.Distribution{{ID: "d1", AppID: "app-cache", SourceType: domain.SourceBNHosted, SigningCertSHA256: "cert"}}
	seedApp(repo, "cache", "en-US", locs, dists)

	_, err := svc.GetAppDetail(context.Background(), "cache", "en-US")
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	repo.mu.Lock()
	delete(repo.apps, "cache")
	repo.mu.Unlock()

	detail, err := svc.GetAppDetail(context.Background(), "cache", "en-US")
	if err != nil {
		t.Fatalf("cache hit failed: %v", err)
	}
	if detail.App.CanonicalSlug != "cache" {
		t.Fatalf("unexpected slug: %s", detail.App.CanonicalSlug)
	}
	b, _ := cache.Get(context.Background(), "catalog:app:cache:en-US")
	var d domain.AppDetail
	if err := json.Unmarshal(b, &d); err != nil {
		t.Fatalf("cached json invalid: %v", err)
	}
}

func TestSanitizeHTML(t *testing.T) {
	svc, _, _ := setupCatalog(t)
	out := svc.SanitizeHTML(`<script>alert(1)</script><b>bold</b>`)
	if strings.Contains(out, "script") {
		t.Fatalf("script survived: %s", out)
	}
}

func TestLocaleFallbackChain(t *testing.T) {
	chain := application.LocaleFallbackChain("ko-KR", "en-US")
	if len(chain) == 0 {
		t.Fatal("empty chain")
	}
	if chain[0] != "ko-KR" {
		t.Fatalf("first should be requested, got %s", chain[0])
	}
	// ensure default locale present
	found := false
	for _, c := range chain {
		if c == "en-US" {
			found = true
		}
	}
	if !found {
		t.Fatalf("default locale missing from chain: %v", chain)
	}
}
