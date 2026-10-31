package application

import (
	"context"
	"testing"
)

func TestVirusTotalClient_FailClosedWithoutKey(t *testing.T) {
	ctx := context.Background()

	// 1. Missing Key must FAIL-CLOSED
	unconfiguredClient := NewVirusTotalClient("")
	_, err := unconfiguredClient.QueryFileReport(ctx, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
	if err == nil {
		t.Fatal("SECURITY VULNERABILITY: VirusTotal client passed clean with unconfigured empty API key! Expected Fail-Closed rejection.")
	}
	if err != ErrVTKeyMissing {
		t.Fatalf("expected ErrVTKeyMissing, got: %v", err)
	}
	t.Logf("Verified Fail-Closed security behavior: %v", err)

	// 2. Explicit Hermetic Testing Key
	testClient := NewVirusTotalClient("test_sandbox_hermetic_clean")
	res, err := testClient.QueryFileReport(ctx, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
	if err != nil {
		t.Fatalf("hermetic test failed: %v", err)
	}
	if !res.IsClean {
		t.Fatal("expected hermetic test to return clean")
	}
}

