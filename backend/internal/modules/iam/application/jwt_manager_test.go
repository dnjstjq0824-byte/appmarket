package application

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func TestParseRSAPrivateKeyPEM_RoundTrip(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}
	for _, tc := range []struct {
		name      string
		blockType string
		der       []byte
	}{
		{"pkcs8", "PRIVATE KEY", mustPKCS8(t, key)},
		{"pkcs1", "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(key)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pemBytes := pem.EncodeToMemory(&pem.Block{Type: tc.blockType, Bytes: tc.der})
			parsed, err := ParseRSAPrivateKeyPEM(pemBytes)
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}
			if parsed.D.Cmp(key.D) != 0 {
				t.Fatalf("parsed key differs from generated key")
			}
			mgr, err := NewJWTManager(parsed, "https://auth.appmarket.local")
			if err != nil {
				t.Fatalf("manager init failed: %v", err)
			}
			if mgr == nil {
				t.Fatalf("expected non-nil manager")
			}
		})
	}
}

func mustPKCS8(t *testing.T, key *rsa.PrivateKey) []byte {
	t.Helper()
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("failed to marshal PKCS#8: %v", err)
	}
	return der
}

func TestParseRSAPrivateKeyPEM_RejectsGarbage(t *testing.T) {
	for _, bad := range []string{"", "not-pem", "-----BEGIN CERTIFICATE-----\nZm9v\n-----END CERTIFICATE-----\n"} {
		if _, err := ParseRSAPrivateKeyPEM([]byte(bad)); err == nil {
			t.Fatalf("expected error for %q", bad)
		}
	}
}

func TestJWTManager_RotationKeepsPreviousTokensValid(t *testing.T) {
	mgr, err := NewJWTManager(nil, "https://auth.appmarket.local")
	if err != nil {
		t.Fatalf("manager init failed: %v", err)
	}
	before, err := mgr.GenerateAccessToken("user-1", "user@test.local", []string{"USER"}, nil, false)
	if err != nil {
		t.Fatalf("issue before rotation failed: %v", err)
	}
	firstKID := mgr.KeyID()

	rotated, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rotation key generation failed: %v", err)
	}
	if err := mgr.RotateKeys(rotated); err != nil {
		t.Fatalf("rotation failed: %v", err)
	}
	if mgr.KeyID() == firstKID {
		t.Fatalf("expected new kid after rotation")
	}

	if _, err := mgr.ValidateAccessToken(before); err != nil {
		t.Fatalf("pre-rotation token must validate during grace: %v", err)
	}
	after, err := mgr.GenerateAccessToken("user-1", "user@test.local", []string{"USER"}, nil, false)
	if err != nil {
		t.Fatalf("issue after rotation failed: %v", err)
	}
	if _, err := mgr.ValidateAccessToken(after); err != nil {
		t.Fatalf("post-rotation token must validate: %v", err)
	}
}

func TestJWTManager_DoubleRotationEvictsOldestKey(t *testing.T) {
	mgr, err := NewJWTManager(nil, "https://auth.appmarket.local")
	if err != nil {
		t.Fatalf("manager init failed: %v", err)
	}
	genA, err := mgr.GenerateAccessToken("user-1", "user@test.local", []string{"USER"}, nil, false)
	if err != nil {
		t.Fatalf("gen-A issue failed: %v", err)
	}
	if err := mgr.RotateKeys(nil); err != nil {
		t.Fatalf("first rotation failed: %v", err)
	}
	genB, err := mgr.GenerateAccessToken("user-1", "user@test.local", []string{"USER"}, nil, false)
	if err != nil {
		t.Fatalf("gen-B issue failed: %v", err)
	}
	if err := mgr.RotateKeys(nil); err != nil {
		t.Fatalf("second rotation failed: %v", err)
	}

	if _, err := mgr.ValidateAccessToken(genA); err == nil {
		t.Fatalf("SECURITY FLAW: token from evicted key generation still validates")
	}
	if _, err := mgr.ValidateAccessToken(genB); err != nil {
		t.Fatalf("previous-generation token must validate during grace: %v", err)
	}
	foreign, err := NewJWTManager(nil, "https://auth.appmarket.local")
	if err != nil {
		t.Fatalf("foreign manager init failed: %v", err)
	}
	foreignToken, err := foreign.GenerateAccessToken("user-1", "user@test.local", []string{"USER"}, nil, false)
	if err != nil {
		t.Fatalf("foreign issue failed: %v", err)
	}
	if _, err := mgr.ValidateAccessToken(foreignToken); err == nil {
		t.Fatalf("SECURITY FLAW: token from unknown key was accepted")
	}
}

func TestJWTManager_JWKSPublishesGraceKey(t *testing.T) {
	mgr, err := NewJWTManager(nil, "https://auth.appmarket.local")
	if err != nil {
		t.Fatalf("manager init failed: %v", err)
	}
	single, ok := mgr.GetJWKS()["keys"].([]map[string]any)
	if !ok || len(single) != 1 {
		t.Fatalf("expected 1 JWKS key before rotation, got %+v", mgr.GetJWKS())
	}
	firstKID := mgr.KeyID()
	if err := mgr.RotateKeys(nil); err != nil {
		t.Fatalf("rotation failed: %v", err)
	}
	dual, ok := mgr.GetJWKS()["keys"].([]map[string]any)
	if !ok || len(dual) != 2 {
		t.Fatalf("expected 2 JWKS keys after rotation, got %+v", mgr.GetJWKS())
	}
	kids := map[string]bool{dual[0]["kid"].(string): true, dual[1]["kid"].(string): true}
	if !kids[mgr.KeyID()] || !kids[firstKID] {
		t.Fatalf("JWKS kids %v must contain current %q and previous %q", kids, mgr.KeyID(), firstKID)
	}
}
