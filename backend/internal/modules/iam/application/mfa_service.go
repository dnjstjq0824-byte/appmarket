package application

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/iam/domain"
)

type MFAService struct {
	issuer string
	period int64
	digits int
}

func NewMFAService(issuer string) *MFAService {
	return &MFAService{
		issuer: issuer,
		period: 30,
		digits: 6,
	}
}

// GenerateSecret generates a secure base32-encoded TOTP secret and recovery codes.
func (s *MFAService) GenerateSecret(email string) (secretBase32 string, qrURL string, recoveryCodes []string, hashedCodes []string, err error) {
	rawSecret := make([]byte, 20)
	if _, err := rand.Read(rawSecret); err != nil {
		return "", "", nil, nil, fmt.Errorf("failed to generate random secret: %w", err)
	}

	secretBase32 = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(rawSecret)

	// Build otpauth URL
	encodedIssuer := url.QueryEscape(s.issuer)
	label := fmt.Sprintf("%s:%s", encodedIssuer, email)
	qrURL = fmt.Sprintf("otpauth://totp/%s?secret=%s&issuer=%s&period=%d&digits=%d",
		label, secretBase32, encodedIssuer, s.period, s.digits)

	// Generate 10 recovery codes
	recoveryCodes = make([]string, 10)
	hashedCodes = make([]string, 10)
	for i := 0; i < 10; i++ {
		codeBytes := make([]byte, 5)
		if _, err := rand.Read(codeBytes); err != nil {
			return "", "", nil, nil, fmt.Errorf("failed to generate recovery code: %w", err)
		}
		code := fmt.Sprintf("%X", codeBytes) // 10 chars hex
		recoveryCodes[i] = code
		hash := sha256.Sum256([]byte(code))
		hashedCodes[i] = fmt.Sprintf("%x", hash)
	}

	return secretBase32, qrURL, recoveryCodes, hashedCodes, nil
}

// ComputeTOTP calculates the TOTP code for a given timestamp.
func (s *MFAService) ComputeTOTP(secretBase32 string, t time.Time) (string, error) {
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(strings.TrimSpace(secretBase32)))
	if err != nil {
		return "", fmt.Errorf("invalid base32 secret: %w", err)
	}

	counter := uint64(t.Unix() / s.period)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)

	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	h := mac.Sum(nil)

	offset := h[len(h)-1] & 0x0f
	binaryCode := binary.BigEndian.Uint32(h[offset:offset+4]) & 0x7fffffff
	otp := binaryCode % 1000000

	return fmt.Sprintf("%06d", otp), nil
}

// VerifyTOTP validates a 6-digit TOTP code with -1, 0, +1 time window drift tolerance.
func (s *MFAService) VerifyTOTP(secretBase32, code string, now time.Time) bool {
	code = strings.TrimSpace(code)
	if len(code) != s.digits {
		return false
	}

	for _, offset := range []int64{-1, 0, 1} {
		t := now.Add(time.Duration(offset*s.period) * time.Second)
		expected, err := s.ComputeTOTP(secretBase32, t)
		if err == nil && expected == code {
			return true
		}
	}
	return false
}

// VerifyAndConsumeRecoveryCode checks if a recovery code is valid and consumes it from the list.
func (s *MFAService) VerifyAndConsumeRecoveryCode(code string, hashedCodes []string) (bool, []string) {
	code = strings.ToUpper(strings.TrimSpace(code))
	targetHash := fmt.Sprintf("%x", sha256.Sum256([]byte(code)))

	for i, h := range hashedCodes {
		if h == targetHash {
			// Remove the consumed code
			remaining := append(hashedCodes[:i], hashedCodes[i+1:]...)
			return true, remaining
		}
	}
	return false, hashedCodes
}

// ValidateMFA checks either a TOTP code or a recovery code.
func (s *MFAService) ValidateMFA(cred *domain.MFACredential, code string, now time.Time) (bool, error) {
	if cred == nil || !cred.IsVerified {
		return false, domain.ErrMFARequired
	}

	// First try as TOTP
	if len(strings.TrimSpace(code)) == s.digits {
		if s.VerifyTOTP(cred.SecretEncrypted, code, now) {
			return true, nil
		}
	}

	// Next try as recovery code
	valid, remaining := s.VerifyAndConsumeRecoveryCode(code, cred.BackupCodesHashed)
	if valid {
		cred.BackupCodesHashed = remaining
		return true, nil
	}

	return false, domain.ErrInvalidMFACode
}
