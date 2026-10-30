package treescan

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidGrantType       = errors.New("invalid grant_type: must be urn:ietf:params:oauth:grant-type:token-exchange")
	ErrInvalidTokenType       = errors.New("invalid subject_token_type: must be urn:ietf:params:oauth:token-type:id_token")
	ErrInvalidIssuer          = errors.New("invalid token issuer: expected https://auth.treescan.internal")
	ErrInvalidAudience        = errors.New("invalid token audience: must contain appmarket-api")
	ErrTokenExpired           = errors.New("subject token has expired")
	ErrMissingSubjectToken    = errors.New("missing subject_token parameter")
)

const (
	RFC8693GrantType      = "urn:ietf:params:oauth:grant-type:token-exchange"
	RFC8693IdTokenType    = "urn:ietf:params:oauth:token-type:id_token"
	ExpectedIssuer        = "https://auth.treescan.internal"
	ExpectedAudience      = "appmarket-api"
)

type TokenExchangeRequest struct {
	GrantType        string `json:"grant_type"`
	SubjectToken     string `json:"subject_token"`
	SubjectTokenType string `json:"subject_token_type"`
	Audience         string `json:"audience,omitempty"`
	Scope            string `json:"scope,omitempty"`
}

type TokenExchangeResponse struct {
	AccessToken     string `json:"access_token"`
	IssuedTokenType string `json:"issued_token_type"`
	TokenType       string `json:"token_type"`
	ExpiresIn       int64  `json:"expires_in"`
	RefreshToken    string `json:"refresh_token"`
	UserID          string `json:"user_id"`
	Email           string `json:"email"`
}

type TreeScanClaims struct {
	Issuer    string
	Subject   string
	Email     string
	Audience  string
	ExpiresAt time.Time
}

type SSOTokenExchangeService struct {
	mu            sync.RWMutex
	adapter       *AdapterHandler
	localUsers    map[string]string // treeScanSub -> localUserID
	signingSecret string
}

func NewSSOTokenExchangeService(adapter *AdapterHandler, secret string) *SSOTokenExchangeService {
	if secret == "" {
		secret = "appmarket_jwt_secret_signing_key_2026"
	}
	return &SSOTokenExchangeService{
		adapter:       adapter,
		localUsers:    make(map[string]string),
		signingSecret: secret,
	}
}

// ParseAndVerifyTreeScanToken mocks or validates the JWT token structure
func (s *SSOTokenExchangeService) parseTreeScanToken(rawToken string) (*TreeScanClaims, error) {
	// Parse simulated or structured token: "iss:sub:email:aud:exp"
	parts := strings.Split(rawToken, "|")
	if len(parts) < 5 {
		return nil, errors.New("malformed subject_token payload")
	}

	expUnix := int64(0)
	fmt.Sscanf(parts[4], "%d", &expUnix)

	claims := &TreeScanClaims{
		Issuer:    parts[0],
		Subject:   parts[1],
		Email:     parts[2],
		Audience:  parts[3],
		ExpiresAt: time.Unix(expUnix, 0),
	}

	return claims, nil
}

// ExchangeToken processes RFC 8693 token exchange from TreeScan to AppMarket session
func (s *SSOTokenExchangeService) ExchangeToken(ctx context.Context, req TokenExchangeRequest) (*TokenExchangeResponse, error) {
	// 1. Independence Feature Flag Check
	if !s.adapter.IsEnabled() {
		return nil, ErrFeatureDisabled
	}

	// 2. RFC 8693 Parameter Validation
	if req.GrantType != RFC8693GrantType {
		return nil, ErrInvalidGrantType
	}
	if req.SubjectTokenType != RFC8693IdTokenType {
		return nil, ErrInvalidTokenType
	}
	if req.SubjectToken == "" {
		return nil, ErrMissingSubjectToken
	}

	// 3. Claims Verification
	claims, err := s.parseTreeScanToken(req.SubjectToken)
	if err != nil {
		return nil, err
	}

	if claims.Issuer != ExpectedIssuer {
		return nil, ErrInvalidIssuer
	}
	if !strings.Contains(claims.Audience, ExpectedAudience) {
		return nil, ErrInvalidAudience
	}
	if time.Now().After(claims.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 4. JIT Provisioning / Account Linking
	localUserID, exists := s.localUsers[claims.Subject]
	if !exists {
		localUserID = fmt.Sprintf("usr-jit-ts-%s", claims.Subject)
		s.localUsers[claims.Subject] = localUserID
	}

	// 5. Issue AppMarket Access & Refresh Tokens
	now := time.Now().UTC()
	accessToken := fmt.Sprintf("am_at_%s_%d", localUserID, now.Unix())
	refreshToken := fmt.Sprintf("am_rt_%s_%d", localUserID, now.Unix())

	return &TokenExchangeResponse{
		AccessToken:     accessToken,
		IssuedTokenType: "urn:ietf:params:oauth:token-type:access_token",
		TokenType:       "Bearer",
		ExpiresIn:       3600, // 1 hour
		RefreshToken:    refreshToken,
		UserID:          localUserID,
		Email:           claims.Email,
	}, nil
}

