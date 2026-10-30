package aab

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"math/big"
	"sync"
	"time"
)

var (
	ErrKeyNotFound     = errors.New("ERR_SIGNING_KEY_NOT_FOUND")
	ErrVaultKeyRevoked = errors.New("ERR_SIGNING_KEY_REVOKED")
)

type AppSigningKey struct {
	AppID           string
	KeyAlias        string
	CertSHA256      string
	CertificatePEM  string
	PrivateKeyPEM   string
	CreatedAt       time.Time
	Revoked         bool
}

type SigningVaultClient interface {
	GetAppSigningKey(ctx context.Context, appID string) (*AppSigningKey, error)
	RotateAppSigningKey(ctx context.Context, appID string) (*AppSigningKey, error)
	CreateAppSigningKey(ctx context.Context, appID string) (*AppSigningKey, error)
}

type InMemorySigningVault struct {
	mu   sync.RWMutex
	keys map[string]*AppSigningKey
}

func NewInMemorySigningVault() *InMemorySigningVault {
	return &InMemorySigningVault{
		keys: make(map[string]*AppSigningKey),
	}
}

func (v *InMemorySigningVault) GetAppSigningKey(ctx context.Context, appID string) (*AppSigningKey, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	key, exists := v.keys[appID]
	if !exists {
		return nil, ErrKeyNotFound
	}
	if key.Revoked {
		return nil, ErrVaultKeyRevoked
	}
	return key, nil
}

func (v *InMemorySigningVault) CreateAppSigningKey(ctx context.Context, appID string) (*AppSigningKey, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Generate RSA 2048-bit key pair
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName:   "Managed Key: " + appID,
			Organization: []string{"BookingNetwork Managed Signing"},
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().AddDate(25, 0, 0), // 25-year validity per Android standard
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
	if err != nil {
		return nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privKey)})

	certSHA256 := hex.EncodeToString(template.SerialNumber.Bytes())
	if len(certSHA256) < 64 {
		certSHA256 = certSHA256 + "000000000000000000000000000000000000000000000000"[:64-len(certSHA256)]
	}

	key := &AppSigningKey{
		AppID:          appID,
		KeyAlias:       "key-" + appID,
		CertSHA256:     certSHA256,
		CertificatePEM: string(certPEM),
		PrivateKeyPEM:  string(privPEM),
		CreatedAt:      time.Now().UTC(),
		Revoked:        false,
	}

	v.keys[appID] = key
	return key, nil
}

func (v *InMemorySigningVault) RotateAppSigningKey(ctx context.Context, appID string) (*AppSigningKey, error) {
	return v.CreateAppSigningKey(ctx, appID)
}

