package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

var (
	ErrInvalidKeySize    = errors.New("crypto: invalid key size, must be 16, 24, or 32 bytes")
	ErrCiphertextTooShort = errors.New("crypto: ciphertext too short")
	ErrDecryptionFailed  = errors.New("crypto: message authentication/decryption failed")
)

// Encrypt encrypts plaintext using AES-GCM with a prepended cryptographic nonce.
// Returns a standard base64-encoded string.
func Encrypt(plaintext []byte, key []byte) (string, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return "", ErrInvalidKeySize
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate secure nonce: %w", err)
	}

	// Seal appends the ciphertext to nonce, so nonce is prepended at the head
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// EncryptString is a convenience helper to encrypt string plaintext.
func EncryptString(plaintext string, key []byte) (string, error) {
	return Encrypt([]byte(plaintext), key)
}

// Decrypt decodes base64 ciphertext, extracts the prepended nonce, and decrypts with AES-GCM.
func Decrypt(encodedCiphertext string, key []byte) ([]byte, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, ErrInvalidKeySize
	}

	data, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 ciphertext: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create gcm: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, ErrCiphertextTooShort
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// DecryptString decrypts and returns string plaintext.
func DecryptString(encodedCiphertext string, key []byte) (string, error) {
	plaintextBytes, err := Decrypt(encodedCiphertext, key)
	if err != nil {
		return "", err
	}
	return string(plaintextBytes), nil
}
