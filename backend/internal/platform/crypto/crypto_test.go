package crypto

import (
	"bytes"
	"errors"
	"testing"
)

// TC-SEC-002: KMS AES-256-GCM Envelope Encryption & Blind Indexing
func TestTCSEC002KMSEnvelopeEncryptionAndBlindIndex(t *testing.T) {
	masterKeyHex := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	blindKeyHex := "fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210"

	enc, err := NewEnvelopeEncryptor(masterKeyHex, blindKeyHex)
	if err != nil {
		t.Fatalf("failed to create encryptor: %v", err)
	}

	sensitiveBankAccount := "110-394-829104-KR-SHINHAN"

	t.Run("Sub-test 1: Encrypt and Decrypt Round-Trip Fidelity", func(t *testing.T) {
		ciphertext, err := enc.EncryptField([]byte(sensitiveBankAccount))
		if err != nil {
			t.Fatalf("encryption failed: %v", err)
		}

		// Ensure ciphertext does NOT contain plaintext
		if bytes.Contains(ciphertext, []byte(sensitiveBankAccount)) {
			t.Fatalf("plaintext leaked inside ciphertext!")
		}

		decrypted, err := enc.DecryptField(ciphertext)
		if err != nil {
			t.Fatalf("decryption failed: %v", err)
		}

		if string(decrypted) != sensitiveBankAccount {
			t.Fatalf("decrypted value mismatch: expected %s, got %s", sensitiveBankAccount, string(decrypted))
		}
	})

	t.Run("Sub-test 2: Tampered Ciphertext Rejection", func(t *testing.T) {
		ciphertext, err := enc.EncryptField([]byte(sensitiveBankAccount))
		if err != nil {
			t.Fatalf("encryption failed: %v", err)
		}

		// Tamper with last byte
		ciphertext[len(ciphertext)-1] ^= 0xFF

		_, err = enc.DecryptField(ciphertext)
		if err == nil || !errors.Is(err, ErrDecryptionFailed) {
			t.Fatalf("expected ErrDecryptionFailed on tampered ciphertext, got %v", err)
		}
	})

	t.Run("Sub-test 3: Deterministic Blind Index for Exact Match Search without Decryption", func(t *testing.T) {
		index1 := enc.ComputeBlindIndex(sensitiveBankAccount)
		index2 := enc.ComputeBlindIndex(sensitiveBankAccount)
		index3 := enc.ComputeBlindIndex("DIFFERENT_ACCOUNT")

		if index1 != index2 {
			t.Fatalf("blind index must be deterministic for same plaintext")
		}
		if index1 == index3 {
			t.Fatalf("different plaintexts must not have identical blind index")
		}
		if len(index1) != 64 {
			t.Fatalf("expected 64-char sha256 blind index, got %d", len(index1))
		}
	})
}
