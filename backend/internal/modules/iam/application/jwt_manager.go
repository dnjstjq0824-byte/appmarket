package application

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
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
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	keyID      string
	issuer     string
	ttl        time.Duration
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
			Issuer:    m.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = m.keyID

	signed, err := token.SignedString(m.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return signed, nil
}

func (m *JWTManager) ValidateAccessToken(tokenStr string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AccessClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.publicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func (m *JWTManager) PublicKey() *rsa.PublicKey {
	return m.publicKey
}

func (m *JWTManager) KeyID() string {
	return m.keyID
}

// GetJWKS returns RFC 7517 compliant JSON Web Key Set representation of public keys.
func (m *JWTManager) GetJWKS() map[string]any {
	nBytes := m.publicKey.N.Bytes()
	eInt := big.NewInt(int64(m.publicKey.E)).Bytes()

	return map[string]any{
		"keys": []map[string]any{
			{
				"kty": "RSA",
				"use": "sig",
				"alg": "RS256",
				"kid": m.keyID,
				"n":   base64.RawURLEncoding.EncodeToString(nBytes),
				"e":   base64.RawURLEncoding.EncodeToString(eInt),
			},
		},
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

