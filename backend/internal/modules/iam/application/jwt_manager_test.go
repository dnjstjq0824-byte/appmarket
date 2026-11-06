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
