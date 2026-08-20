package utils_test

import (
	"testing"

	"github.com/shankywho/ropus/backend/internal/utils"
)

func TestEncryptDecrypt(t *testing.T) {
	kms := utils.NewMockKMS()
	tenantID := "tenant_fintech_01"

	key, err := kms.GetTenantKey(tenantID)
	if err != nil {
		t.Fatalf("failed to get tenant key: %v", err)
	}

	plaintext := "192.168.1.100|device_fp_secret_9988"

	// 1. Encrypt
	ciphertext, err := utils.EncryptString(plaintext, key)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	if ciphertext == plaintext || len(ciphertext) == 0 {
		t.Errorf("ciphertext is invalid or matches plaintext")
	}

	// 2. Nonce Uniqueness Check (two encryptions of same text produce distinct ciphertexts)
	ciphertext2, _ := utils.EncryptString(plaintext, key)
	if ciphertext == ciphertext2 {
		t.Errorf("expected distinct ciphertexts due to randomized nonces")
	}

	// 3. Decrypt
	decrypted, err := utils.DecryptString(ciphertext, key)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("expected decrypted '%s', got '%s'", plaintext, decrypted)
	}

	// 4. Decrypt with Wrong Key (should fail)
	otherKey, _ := kms.GetTenantKey("tenant_fintech_02")
	_, err = utils.DecryptString(ciphertext, otherKey)
	if err == nil {
		t.Errorf("expected decryption to fail with wrong tenant key")
	}
}

func TestMockKMS_CryptoShredding(t *testing.T) {
	kms := utils.NewMockKMS()
	tenantID := "tenant_gdpr_shred_test"

	key, err := kms.GetTenantKey(tenantID)
	if err != nil {
		t.Fatalf("failed to get tenant key: %v", err)
	}

	sensitivePII := "user_card_token_and_ip"
	ciphertext, err := utils.EncryptString(sensitivePII, key)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Verify decrypt works before shredding
	decrypted, err := utils.DecryptString(ciphertext, key)
	if err != nil || decrypted != sensitivePII {
		t.Fatalf("pre-shred decrypt failed: %v", err)
	}

	// Execute Crypto-Shredding
	if err := kms.ShredTenantKey(tenantID); err != nil {
		t.Fatalf("shredding key failed: %v", err)
	}

	if !kms.IsTenantShredded(tenantID) {
		t.Errorf("expected tenant to be marked as shredded")
	}

	// Attempting to get key after shredding must fail
	_, err = kms.GetTenantKey(tenantID)
	if err != utils.ErrTenantKeyShredded {
		t.Errorf("expected ErrTenantKeyShredded, got %v", err)
	}
}
