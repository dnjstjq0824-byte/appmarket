package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

var (
	ErrDecryptionFailed  = errors.New("ERR_CIPHERTEXT_DECRYPTION_FAILED")
	ErrCiphertextTooShort = errors.New("ERR_CIPHERTEXT_TOO_SHORT")
)

type EnvelopeEncryptor struct {
	masterKey      []byte // 256-bit (32 bytes) master key
	blindIndexKey  []byte // 256-bit (32 bytes) blind index key
}

func NewEnvelopeEncryptor(masterKeyHex, blindIndexKeyHex string) (*EnvelopeEncryptor, error) {
	mk, err := hex.DecodeString(masterKeyHex)
	if err != nil || len(mk) != 32 {
		return nil, errors.New("master key must be a valid 64-hex (32 bytes) string")
	}

	bk, err := hex.DecodeString(blindIndexKeyHex)
	if err != nil || len(bk) != 32 {
		return nil, errors.New("blind index key must be a valid 64-hex (32 bytes) string")
	}

	return &EnvelopeEncryptor{
		masterKey:     mk,
		blindIndexKey: bk,
	}, nil
}

// EncryptField encrypts plaintext using AES-256-GCM with standard 12-byte random nonce
func (e *EnvelopeEncryptor) EncryptField(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.masterKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Ciphertext contains: nonce + ciphertext + GCM auth tag
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// DecryptField decrypts AES-256-GCM ciphertext and validates integrity
func (e *EnvelopeEncryptor) DecryptField(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.masterKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrCiphertextTooShort
	}

	nonce := ciphertext[:nonceSize]
	actualCiphertext := ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	return plaintext, nil
}

// ComputeBlindIndex creates a one-way deterministic HMAC search hash for encrypted fields
func (e *EnvelopeEncryptor) ComputeBlindIndex(plaintext string) string {
	mac := hmac.New(sha256.New, e.blindIndexKey)
	mac.Write([]byte(plaintext))
	return hex.EncodeToString(mac.Sum(nil))
}
