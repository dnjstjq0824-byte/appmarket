package application

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/iam/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/sanitize"
)

// IdentityRepository abstracts federated identity persistence.
type IdentityRepository interface {
	FindUserIDByProvider(ctx context.Context, provider, subject string) (string, error)
	LinkIdentity(ctx context.Context, identity *domain.UserIdentity) error
}

// OAuthProviderConfig describes one social login provider. Endpoint URLs are
// overridable so tests can point at a mock OIDC server.
type OAuthProviderConfig struct {
	Provider      string
	ClientID      string
	ClientSecret  string
	AuthURL       string
	TokenURL      string
	JWKURL        string
	UserInfoURL   string
	UserEmailsURL string
	Issuer        string
	Scopes        []string
	// Apple Sign In uses an ES256 client secret instead of a static secret.
	AppleTeamID       string
	AppleKeyID        string
	ApplePrivateKeyPEM string
}

// DefaultOAuthProviders returns production endpoint presets. A provider is
// active only when its client ID is configured.
func DefaultOAuthProviders() map[string]OAuthProviderConfig {
	return map[string]OAuthProviderConfig{
		"google": {
			Provider: domain.IdentityProviderGoogle,
			AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
			JWKURL:   "https://www.googleapis.com/oauth2/v3/certs",
			Issuer:   "https://accounts.google.com",
			Scopes:   []string{"openid", "email", "profile"},
		},
		"github": {
			Provider:      domain.IdentityProviderGitHub,
			AuthURL:       "https://github.com/login/oauth/authorize",
			TokenURL:      "https://github.com/login/oauth/access_token",
			UserInfoURL:   "https://api.github.com/user",
			UserEmailsURL: "https://api.github.com/user/emails",
			Scopes:        []string{"read:user", "user:email"},
		},
		"apple": {
			Provider: domain.IdentityProviderApple,
			AuthURL:  "https://appleid.apple.com/auth/authorize",
			TokenURL: "https://appleid.apple.com/auth/token",
			JWKURL:   "https://appleid.apple.com/auth/keys",
			Issuer:   "https://appleid.apple.com",
			Scopes:   []string{"openid", "email", "name"},
		},
	}
}

const (
	oauthStateTTL     = 10 * time.Minute
	oauthHTTPTimeout  = 10 * time.Second
	oauthSecretLeeway = 5 * time.Minute
)

type OAuthService struct {
	auth         *AuthService
	users        UserRepository
	identities   IdentityRepository
	http         *http.Client
	redirectBase string

	mu        sync.RWMutex
	providers map[string]*OAuthProviderConfig
	states    map[string]time.Time
}

func NewOAuthService(auth *AuthService, users UserRepository, identities IdentityRepository, redirectBaseURL string) *OAuthService {
	return &OAuthService{
		auth:         auth,
		users:        users,
		identities:   identities,
		http:         &http.Client{Timeout: oauthHTTPTimeout},
		redirectBase: strings.TrimRight(redirectBaseURL, "/"),
		providers:    make(map[string]*OAuthProviderConfig),
		states:       make(map[string]time.Time),
	}
}

func (s *OAuthService) RegisterProvider(cfg OAuthProviderConfig) {
	if cfg.ClientID == "" {
		return
	}
	short := ""
	for name, preset := range DefaultOAuthProviders() {
		if preset.Provider == cfg.Provider {
			short = name
			break
		}
	}
	if short == "" {
		return
	}
	if cfg.AuthURL == "" || cfg.TokenURL == "" {
		preset := DefaultOAuthProviders()[short]
		if cfg.AuthURL == "" {
			cfg.AuthURL = preset.AuthURL
		}
		if cfg.TokenURL == "" {
			cfg.TokenURL = preset.TokenURL
		}
		if cfg.JWKURL == "" {
			cfg.JWKURL = preset.JWKURL
		}
		if cfg.UserInfoURL == "" {
			cfg.UserInfoURL = preset.UserInfoURL
		}
		if cfg.UserEmailsURL == "" {
			cfg.UserEmailsURL = preset.UserEmailsURL
		}
		if cfg.Issuer == "" {
			cfg.Issuer = preset.Issuer
		}
		if len(cfg.Scopes) == 0 {
			cfg.Scopes = preset.Scopes
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := cfg
	s.providers[short] = &cp
}

// Providers lists active provider short names (google, github, apple).
func (s *OAuthService) Providers() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	names := make([]string, 0, len(s.providers))
	for name := range s.providers {
		names = append(names, name)
	}
	return names
}

func (s *OAuthService) provider(short string) (*OAuthProviderConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cfg, ok := s.providers[strings.ToLower(short)]
	if !ok {
		return nil, domain.ErrOAuthProviderUnknown
	}
	return cfg, nil
}

func (s *OAuthService) redirectURI(short string) string {
	return s.redirectBase + "/" + strings.ToLower(short) + "/callback"
}

// GenerateState mints a single-use CSRF token for the login flow.
func (s *OAuthService) GenerateState() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	state := hex.EncodeToString(b[:])
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[state] = time.Now().Add(oauthStateTTL)
	return state, nil
}

func (s *OAuthService) consumeState(state string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	exp, ok := s.states[state]
	if !ok {
		return false
	}
	delete(s.states, state)
	return time.Now().Before(exp)
}

// LoginURL builds the provider authorization redirect for a fresh state.
func (s *OAuthService) LoginURL(short, state string) (string, error) {
	cfg, err := s.provider(short)
	if err != nil {
		return "", err
	}
	q := url.Values{}
	q.Set("client_id", cfg.ClientID)
	q.Set("redirect_uri", s.redirectURI(short))
	q.Set("response_type", "code")
	q.Set("scope", strings.Join(cfg.Scopes, " "))
	q.Set("state", state)
	if cfg.Provider == domain.IdentityProviderApple {
		q.Set("response_mode", "query")
	}
	return cfg.AuthURL + "?" + q.Encode(), nil
}

type oauthTokenResponse struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
	TokenType   string `json:"token_type"`
}

type federatedClaims struct {
	Subject       string
	Email         string
	EmailVerified bool
	DisplayName   string
}

// HandleCallback exchanges the authorization code, verifies the provider
// identity, and returns local session tokens (provisioning or linking the
// local user as needed).
func (s *OAuthService) HandleCallback(ctx context.Context, short, code, state, ipAddress, userAgent string) (*AuthResult, error) {
	if !s.consumeState(state) {
		return nil, domain.ErrOAuthStateInvalid
	}
	cfg, err := s.provider(short)
	if err != nil {
		return nil, err
	}
	if code == "" {
		return nil, domain.ErrOAuthExchangeFailed
	}
	token, err := s.exchangeCode(ctx, cfg, s.redirectURI(short), code)
	if err != nil {
		return nil, err
	}
	var claims *federatedClaims
	if cfg.Provider == domain.IdentityProviderGitHub {
		claims, err = s.gitHubClaims(ctx, cfg, token.AccessToken)
	} else {
		claims, err = s.verifyIDToken(ctx, cfg, token.IDToken)
	}
	if err != nil {
		return nil, err
	}
	if claims.Email == "" {
		return nil, domain.ErrOAuthEmailMissing
	}
	return s.loginOrLink(ctx, cfg.Provider, claims, ipAddress, userAgent)
}

func (s *OAuthService) exchangeCode(ctx context.Context, cfg *OAuthProviderConfig, redirectURI, code string) (*oauthTokenResponse, error) {
	secret := cfg.ClientSecret
	if cfg.Provider == domain.IdentityProviderApple {
		generated, err := appleClientSecret(cfg)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", domain.ErrOAuthExchangeFailed, err)
		}
		secret = generated
	}
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	form.Set("client_id", cfg.ClientID)
	form.Set("client_secret", secret)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrOAuthExchangeFailed, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := s.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrOAuthExchangeFailed, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrOAuthExchangeFailed, err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("%w: provider status %d", domain.ErrOAuthExchangeFailed, resp.StatusCode)
	}
	var token oauthTokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrOAuthExchangeFailed, err)
	}
	if token.AccessToken == "" && token.IDToken == "" {
		return nil, domain.ErrOAuthExchangeFailed
	}
	return &token, nil
}

type oidcClaims struct {
	Email         string `json:"email"`
	EmailVerified *bool  `json:"email_verified"`
	Name          string `json:"name"`
	jwt.RegisteredClaims
}

func (s *OAuthService) verifyIDToken(ctx context.Context, cfg *OAuthProviderConfig, rawToken string) (*federatedClaims, error) {
	if rawToken == "" {
		return nil, domain.ErrOAuthIDTokenInvalid
	}
	unverified, _, err := new(jwt.Parser).ParseUnverified(rawToken, &oidcClaims{})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrOAuthIDTokenInvalid, err)
	}
	kid, _ := unverified.Header["kid"].(string)
	key, err := s.jwkPublicKey(ctx, cfg, kid)
	if err != nil {
		return nil, err
	}
	claims := &oidcClaims{}
	parsed, err := jwt.ParseWithClaims(rawToken, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return key, nil
	}, jwt.WithIssuer(cfg.Issuer), jwt.WithAudience(cfg.ClientID), jwt.WithExpirationRequired(), jwt.WithLeeway(oauthSecretLeeway))
	if err != nil || !parsed.Valid {
		return nil, fmt.Errorf("%w: %v", domain.ErrOAuthIDTokenInvalid, err)
	}
	verified := true
	if claims.EmailVerified != nil {
		verified = *claims.EmailVerified
	}
	if claims.Subject == "" {
		return nil, domain.ErrOAuthIDTokenInvalid
	}
	return &federatedClaims{Subject: claims.Subject, Email: strings.ToLower(strings.TrimSpace(claims.Email)), EmailVerified: verified, DisplayName: claims.Name}, nil
}

type jwkSet struct {
	Keys []struct {
		Kid string `json:"kid"`
		Kty string `json:"kty"`
		N   string `json:"n"`
		E   string `json:"e"`
	} `json:"keys"`
}

func (s *OAuthService) jwkPublicKey(ctx context.Context, cfg *OAuthProviderConfig, kid string) (crypto.PublicKey, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.JWKURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrOAuthIDTokenInvalid, err)
	}
	resp, err := s.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrOAuthIDTokenInvalid, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrOAuthIDTokenInvalid, err)
	}
	var set jwkSet
	if err := json.Unmarshal(body, &set); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrOAuthIDTokenInvalid, err)
	}
	for _, k := range set.Keys {
		if k.Kty != "RSA" || (kid != "" && k.Kid != kid) {
			continue
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			continue
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			continue
		}
		e := 0
		for _, b := range eBytes {
			e = e<<8 + int(b)
		}
		return &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: e}, nil
	}
	return nil, fmt.Errorf("%w: signing key %q not found", domain.ErrOAuthIDTokenInvalid, kid)
}

func (s *OAuthService) gitHubClaims(ctx context.Context, cfg *OAuthProviderConfig, accessToken string) (*federatedClaims, error) {
	if accessToken == "" {
		return nil, domain.ErrOAuthIDTokenInvalid
	}
	get := func(targetURL string, out any) error {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Accept", "application/vnd.github+json")
		resp, err := s.http.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			return fmt.Errorf("github status %d", resp.StatusCode)
		}
		return json.NewDecoder(resp.Body).Decode(out)
	}
	var profile struct {
		ID    int64  `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
		Login string `json:"login"`
	}
	if err := get(cfg.UserInfoURL, &profile); err != nil || profile.ID == 0 {
		return nil, fmt.Errorf("%w: %v", domain.ErrOAuthIDTokenInvalid, err)
	}
	email := strings.ToLower(strings.TrimSpace(profile.Email))
	if email == "" {
		var emails []struct {
			Email    string `json:"email"`
			Primary  bool   `json:"primary"`
			Verified bool   `json:"verified"`
		}
		if err := get(cfg.UserEmailsURL, &emails); err != nil {
			return nil, fmt.Errorf("%w: %v", domain.ErrOAuthIDTokenInvalid, err)
		}
		for _, e := range emails {
			if e.Primary && e.Verified {
				email = strings.ToLower(strings.TrimSpace(e.Email))
				break
			}
		}
	}
	name := profile.Name
	if name == "" {
		name = profile.Login
	}
	return &federatedClaims{Subject: fmt.Sprintf("%d", profile.ID), Email: email, EmailVerified: email != "", DisplayName: name}, nil
}

// appleClientSecret builds the ES256-signed client secret Apple requires,
// valid for five minutes from issuance.
func appleClientSecret(cfg *OAuthProviderConfig) (string, error) {
	if cfg.AppleTeamID == "" || cfg.AppleKeyID == "" || cfg.ApplePrivateKeyPEM == "" {
		return "", errors.New("apple team/key configuration missing")
	}
	block, _ := pem.Decode([]byte(cfg.ApplePrivateKeyPEM))
	if block == nil {
		return "", errors.New("no PEM block in apple private key")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse apple private key: %w", err)
	}
	ecKey, ok := parsed.(*ecdsa.PrivateKey)
	if !ok {
		return "", errors.New("apple private key is not ECDSA")
	}
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iss": cfg.AppleTeamID,
		"iat": now.Unix(),
		"exp": now.Add(oauthSecretLeeway).Unix(),
		"aud": "https://appleid.apple.com",
		"sub": cfg.ClientID,
	})
	token.Header["kid"] = cfg.AppleKeyID
	return token.SignedString(ecKey)
}

func (s *OAuthService) loginOrLink(ctx context.Context, provider string, claims *federatedClaims, ipAddress, userAgent string) (*AuthResult, error) {
	email := strings.ToLower(strings.TrimSpace(claims.Email))
	if userID, err := s.identities.FindUserIDByProvider(ctx, provider, claims.Subject); err == nil {
		user, err := s.users.GetUserByID(ctx, userID)
		if err != nil || user == nil {
			return nil, domain.ErrUserNotFound
		}
		return s.issueForUser(ctx, user, ipAddress, userAgent)
	}

	var user *domain.User
	if existing, _ := s.users.GetUserByEmailBlindIndex(ctx, ComputeBlindIndex(email, s.auth.cfg.HMACSecret)); existing != nil {
		user = existing
	} else {
		displayName := sanitize.Strict(strings.TrimSpace(claims.DisplayName))
		if displayName == "" {
			displayName = strings.Split(email, "@")[0]
		}
		now := time.Now().UTC()
		user = &domain.User{
			ID:              uuid.New().String(),
			Email:           email,
			EmailBlindIndex: ComputeBlindIndex(email, s.auth.cfg.HMACSecret),
			PasswordHash:    "!",
			DisplayName:     displayName,
			Status:          domain.StatusActive,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := s.users.CreateUser(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to provision oauth user: %w", err)
		}
	}
	if err := s.identities.LinkIdentity(ctx, &domain.UserIdentity{
		UserID:          user.ID,
		Provider:        provider,
		ProviderSubject: claims.Subject,
		ProviderEmail:   email,
	}); err != nil {
		return nil, fmt.Errorf("failed to link oauth identity: %w", err)
	}
	return s.issueForUser(ctx, user, ipAddress, userAgent)
}

func (s *OAuthService) issueForUser(ctx context.Context, user *domain.User, ipAddress, userAgent string) (*AuthResult, error) {
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now().UTC()) {
		return nil, domain.ErrAccountLocked
	}
	roles, orgRoles, _ := s.users.GetUserRoles(ctx, user.ID)
	if len(roles) == 0 {
		roles = []string{"USER"}
	}
	mfaCred, _ := s.auth.mfaRepo.GetMFACredential(ctx, user.ID)
	if (mfaCred != nil && mfaCred.IsVerified) || isMFAMandatoryForRoles(roles) {
		return &AuthResult{MFARequired: true, User: user}, nil
	}
	return s.auth.createSessionTokens(ctx, user, roles, orgRoles, false, ipAddress, userAgent)
}
