package application

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/iam/domain"
)

func TestAppleClientSecretSignsVerifiableJWT(t *testing.T) {
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ec key: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(ecKey)
	if err != nil {
		t.Fatalf("failed to marshal ec key: %v", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})

	secret, err := appleClientSecret(&OAuthProviderConfig{
		Provider:           domain.IdentityProviderApple,
		ClientID:           "com.example.app",
		AppleTeamID:        "TEAM123456",
		AppleKeyID:         "KEY1234567",
		ApplePrivateKeyPEM: string(pemBytes),
	})
	if err != nil {
		t.Fatalf("secret generation failed: %v", err)
	}
	claims := jwt.MapClaims{}
	parsed, err := jwt.ParseWithClaims(secret, claims, func(tk *jwt.Token) (any, error) {
		if tk.Header["kid"] != "KEY1234567" {
			t.Fatalf("unexpected kid: %v", tk.Header["kid"])
		}
		return &ecKey.PublicKey, nil
	})
	if err != nil || !parsed.Valid {
		t.Fatalf("secret failed verification: %v", err)
	}
	if claims["iss"] != "TEAM123456" || claims["sub"] != "com.example.app" || claims["aud"] != "https://appleid.apple.com" {
		t.Fatalf("unexpected claims: %v", claims)
	}
}

func TestAppleClientSecretRejectsBadConfig(t *testing.T) {
	if _, err := appleClientSecret(&OAuthProviderConfig{Provider: domain.IdentityProviderApple}); err == nil {
		t.Fatalf("expected error for missing team/key config")
	}
	if _, err := appleClientSecret(&OAuthProviderConfig{
		Provider:           domain.IdentityProviderApple,
		ClientID:           "com.example.app",
		AppleTeamID:        "T",
		AppleKeyID:         "K",
		ApplePrivateKeyPEM: "not-pem",
	}); err == nil {
		t.Fatalf("expected error for malformed private key")
	}
}
