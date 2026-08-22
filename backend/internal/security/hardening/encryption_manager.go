package hardening

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"strings"
)

// EncryptionManager handles field-level AES-256 GCM encryption and sanitization.
type EncryptionManager struct {
	key []byte
}

// NewEncryptionManager initializes the encryption manager with a 256-bit key.
func NewEncryptionManager(keyHex string) (*EncryptionManager, error) {
	var key []byte
	if keyHex == "" {
		// Generate random 32-byte key
		key = make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return nil, err
		}
	} else {
		var err error
		key, err = hex.DecodeString(keyHex)
		if err != nil || len(key) != 32 {
			return nil, errors.New("key must be 32 bytes hex encoded (256-bit)")
		}
	}
	return &EncryptionManager{key: key}, nil
}

// EncryptField encrypts plaintext using AES-256 GCM with a unique nonce.
func (e *EncryptionManager) EncryptField(plaintext string) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

// DecryptField decrypts an AES-256 GCM ciphertext.
func (e *EncryptionManager) DecryptField(ciphertextHex string) (string, error) {
	data, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext is too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// SanitizeInput inspects strings for SQL injection or script injection signatures.
func SanitizeInput(input string) (bool, string) {
	lower := strings.ToLower(input)
	patterns := []string{
		"--", "drop table", "union select", "insert into", "delete from",
		"<script", "javascript:", "onload=", "onerror=",
	}

	for _, p := range patterns {
		if strings.Contains(lower, p) {
			return false, "input contains dangerous characters or SQL/XSS tokens"
		}
	}
	return true, ""
}
