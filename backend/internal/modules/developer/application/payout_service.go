package application

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/developer/domain"
)

// PayoutAccountInput carries plaintext payout registration fields. Plaintext
// must never be persisted or returned; only the ciphertext envelope is stored.
type PayoutAccountInput struct {
	Currency      string
	BankName      string
	HolderName    string
	AccountNumber string
	RoutingCode   string
}

// PayoutService encrypts and manages org payout accounts with AES-256-GCM.
type PayoutService struct {
	repo DeveloperRepository
	aead cipher.AEAD
}

// NewPayoutService builds a PayoutService. The key must be exactly 32 bytes.
func NewPayoutService(repo DeveloperRepository, key []byte) (*PayoutService, error) {
	if len(key) != 32 {
		return nil, errors.New("payout encryption key must be 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &PayoutService{repo: repo, aead: aead}, nil
}

// SavePayoutAccount validates, encrypts, and stores an org payout account.
func (s *PayoutService) SavePayoutAccount(ctx context.Context, orgID string, input PayoutAccountInput) (*domain.PayoutAccount, error) {
	if strings.TrimSpace(orgID) == "" {
		return nil, errors.New("org_id is required")
	}
	currency := strings.ToUpper(strings.TrimSpace(input.Currency))
	if len(currency) != 3 || !isASCIIUpper(currency) {
		return nil, fmt.Errorf("currency must be a 3-letter ISO code")
	}
	bank := strings.TrimSpace(input.BankName)
	holder := strings.TrimSpace(input.HolderName)
	if bank == "" || holder == "" {
		return nil, errors.New("bank_name and holder_name are required")
	}
	digits := accountDigits(input.AccountNumber)
	if len(digits) < 6 || len(digits) > 32 {
		return nil, errors.New("account_number must contain 6..32 digits")
	}
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	sealed := s.aead.Seal(nonce, nonce, []byte(digits), []byte(orgID))
	rec := &domain.PayoutRecord{
		OrgID:       orgID,
		Currency:    currency,
		BankName:    bank,
		HolderName:  holder,
		RoutingCode: strings.TrimSpace(input.RoutingCode),
		Last4:       digits[len(digits)-4:],
		Ciphertext:  sealed,
		IsVerified:  false,
	}
	existing, err := s.repo.GetPayoutRecord(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		rec.IsVerified = existing.IsVerified
	}
	if err := s.repo.SavePayoutRecord(ctx, rec); err != nil {
		return nil, err
	}
	return toPayoutAccount(rec), nil
}

// GetPayoutAccount returns the masked payout account view.
func (s *PayoutService) GetPayoutAccount(ctx context.Context, orgID string) (*domain.PayoutAccount, error) {
	rec, err := s.repo.GetPayoutRecord(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, domain.ErrPayoutAccountNotFound
	}
	return toPayoutAccount(rec), nil
}

func toPayoutAccount(rec *domain.PayoutRecord) *domain.PayoutAccount {
	return &domain.PayoutAccount{
		OrgID:        rec.OrgID,
		Currency:     rec.Currency,
		BankName:     rec.BankName,
		HolderName:   rec.HolderName,
		AccountLast4: rec.Last4,
		RoutingCode:  rec.RoutingCode,
		IsVerified:   rec.IsVerified,
	}
}

func accountDigits(raw string) string {
	var b strings.Builder
	for _, r := range raw {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func isASCIIUpper(s string) bool {
	for _, r := range s {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}
