package application

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AccessClaims struct {
	UserID      string            `json:"sub"`
	Email       string            `json:"email"`
	Roles       []string          `json:"roles"`
	OrgRoles    map[string]string `json:"org_roles"`
	MFAVerified bool              `json:"mfa_verified"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	mu            sync.RWMutex
	privateKey    *rsa.PrivateKey
	publicKey     *rsa.PublicKey
	keyID         string
	prevPublicKey *rsa.PublicKey
	prevKeyID     string
	issuer        string
	ttl           time.Duration
}

// ParseRSAPrivateKeyPEM decodes a PKCS#1 ("RSA PRIVATE KEY") or PKCS#8
// ("PRIVATE KEY") PEM block into an RSA signing key.
func ParseRSAPrivateKeyPEM(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("no PEM block found in JWT private key")
	}
	if block.Type == "RSA PRIVATE KEY" {
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS#1 private key: %w", err)
		}
		return key, nil
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PKCS#8 private key: %w", err)
	}
	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("PKCS#8 key is not an RSA private key")
	}
	return key, nil
}

func NewJWTManager(privateKey *rsa.PrivateKey, issuer string) (*JWTManager, error) {
	if privateKey == nil {
		var err error
		privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, fmt.Errorf("failed to generate RSA key: %w", err)
		}
	}

	keyID := uuid.New().String()
	return &JWTManager{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
		keyID:      keyID,
		issuer:     issuer,
		ttl:        15 * time.Minute,
	}, nil
}

func (m *JWTManager) GenerateAccessToken(userID, email string, roles []string, orgRoles map[string]string, mfaVerified bool) (string, error) {
	m.mu.RLock()
	privateKey, keyID, issuer, ttl := m.privateKey, m.keyID, m.issuer, m.ttl
	m.mu.RUnlock()

	now := time.Now().UTC()
	tokenID := uuid.New().String()

	claims := AccessClaims{
		UserID:      userID,
		Email:       email,
		Roles:       roles,
		OrgRoles:    orgRoles,
		MFAVerified: mfaVerified,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			Issuer:    issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = keyID

	signed, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return signed, nil
}

func (m *JWTManager) validateWithKey(tokenStr string, key *rsa.PublicKey) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AccessClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return key, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func (m *JWTManager) ValidateAccessToken(tokenStr string) (*AccessClaims, error) {
	m.mu.RLock()
	current, prev := m.publicKey, m.prevPublicKey
	m.mu.RUnlock()

	claims, err := m.validateWithKey(tokenStr, current)
	if err == nil {
		return claims, nil
	}
	if prev != nil {
		if claims, prevErr := m.validateWithKey(tokenStr, prev); prevErr == nil {
			return claims, nil
		}
	}
	return nil, fmt.Errorf("invalid token: %w", err)
}

// RotateKeys activates newPrivateKey for signing and keeps the previous public
// key for validation grace until the next rotation. A nil key generates a fresh
// 2048-bit RSA key. Previously issued tokens keep validating; only keys older
// than one rotation are rejected.
func (m *JWTManager) RotateKeys(newPrivateKey *rsa.PrivateKey) error {
	var err error
	if newPrivateKey == nil {
		newPrivateKey, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return fmt.Errorf("failed to generate rotation key: %w", err)
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.prevPublicKey = m.publicKey
	m.prevKeyID = m.keyID
	m.privateKey = newPrivateKey
	m.publicKey = &newPrivateKey.PublicKey
	m.keyID = uuid.New().String()
	return nil
}

func (m *JWTManager) PublicKey() *rsa.PublicKey {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.publicKey
}

func (m *JWTManager) KeyID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.keyID
}

// GetJWKS returns RFC 7517 compliant JSON Web Key Set representation of public keys.
// After a rotation it publishes both the current and the previous (grace) key
// so verifiers can validate tokens issued before the rotation.
func (m *JWTManager) GetJWKS() map[string]any {
	m.mu.RLock()
	current, currentKID, prev, prevKID := m.publicKey, m.keyID, m.prevPublicKey, m.prevKeyID
	m.mu.RUnlock()

	keys := []map[string]any{rsaPublicJWK(current, currentKID)}
	if prev != nil {
		keys = append(keys, rsaPublicJWK(prev, prevKID))
	}
	return map[string]any{"keys": keys}
}

func rsaPublicJWK(key *rsa.PublicKey, kid string) map[string]any {
	nBytes := key.N.Bytes()
	eInt := big.NewInt(int64(key.E)).Bytes()
	return map[string]any{
		"kty": "RSA",
		"use": "sig",
		"alg": "RS256",
		"kid": kid,
		"n":   base64.RawURLEncoding.EncodeToString(nBytes),
		"e":   base64.RawURLEncoding.EncodeToString(eInt),
	}
}

// GetOpenIDConfiguration returns RFC 8414 OIDC Discovery configuration metadata.
func (m *JWTManager) GetOpenIDConfiguration(baseURL string) map[string]any {
	return map[string]any{
		"issuer":                                baseURL,
		"authorization_endpoint":                baseURL + "/api/v1/auth/authorize",
		"token_endpoint":                        baseURL + "/api/v1/auth/token",
		"jwks_uri":                              baseURL + "/api/v1/auth/jwks.json",
		"userinfo_endpoint":                     baseURL + "/api/v1/users/me",
		"response_types_supported":              []string{"code", "token", "id_token"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"scopes_supported":                      []string{"openid", "profile", "email"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_post", "client_secret_basic"},
		"claims_supported":                      []string{"sub", "iss", "aud", "exp", "iat", "email", "roles"},
	}
}
