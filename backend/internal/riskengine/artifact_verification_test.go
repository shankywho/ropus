package riskengine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArtifactStore_ImmutabilityAndChecksum(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewLocalFilesystemArtifactStore(tmpDir)
	require.NoError(t, err)

	ctx := context.Background()
	content := "fake_onnx_model_binary_content_v3_1"
	modelID := "model_test_immutability"
	filename := "fraud_model_25f.onnx"

	// 1. Initial Store
	uri, checksum, err := store.StoreArtifact(ctx, modelID, filename, strings.NewReader(content))
	require.NoError(t, err)
	assert.NotEmpty(t, uri)
	assert.NotEmpty(t, checksum)

	// Verify expected SHA-256
	expectedSum := sha256.Sum256([]byte(content))
	assert.Equal(t, hex.EncodeToString(expectedSum[:]), checksum)

	// 2. Existence Check
	assert.True(t, store.ArtifactExists(ctx, uri))

	// 3. Immutability Violation: Attempting to overwrite MUST error
	_, _, errDup := store.StoreArtifact(ctx, modelID, filename, strings.NewReader("different_content"))
	assert.Error(t, errDup)
	assert.Contains(t, errDup.Error(), "immutability violation")
}

func TestArtifactVerifier_ValidArtifact(t *testing.T) {
	tmpDir := t.TempDir()
	verifier := NewArtifactVerifier()
	ctx := context.Background()

	content := "onnx_model_weights_25f_candidate"
	artPath := filepath.Join(tmpDir, "model.onnx")
	err := os.WriteFile(artPath, []byte(content), 0644)
	require.NoError(t, err)

	sum := sha256.Sum256([]byte(content))
	expectedSum := hex.EncodeToString(sum[:])

	record, err := verifier.VerifyArtifact(ctx, "model_01", "v3.1", artPath, expectedSum)
	require.NoError(t, err)
	require.NotNil(t, record)
	assert.True(t, record.Passed)
	assert.Equal(t, uint16(25), record.InputFeaturesCount)
	assert.Equal(t, expectedSum, record.ChecksumSHA256)
	assert.True(t, record.TestVectorPassed)
}

func TestArtifactVerifier_ChecksumMismatch(t *testing.T) {
	tmpDir := t.TempDir()
	verifier := NewArtifactVerifier()
	ctx := context.Background()

	content := "onnx_model_weights_25f_candidate"
	artPath := filepath.Join(tmpDir, "model.onnx")
	err := os.WriteFile(artPath, []byte(content), 0644)
	require.NoError(t, err)

	wrongChecksum := "0000000000000000000000000000000000000000000000000000000000000000"

	record, err := verifier.VerifyArtifact(ctx, "model_01", "v3.1", artPath, wrongChecksum)
	assert.Error(t, err)
	assert.False(t, record.Passed)
	assert.Contains(t, record.Violations[0], "checksum mismatch")
}

func TestArtifactVerifier_EmptyOrMissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	verifier := NewArtifactVerifier()
	ctx := context.Background()

	// Missing file
	_, errMissing := verifier.VerifyArtifact(ctx, "model_01", "v3.1", filepath.Join(tmpDir, "missing.onnx"), "")
	assert.Error(t, errMissing)

	// Empty file
	emptyPath := filepath.Join(tmpDir, "empty.onnx")
	_ = os.WriteFile(emptyPath, []byte(""), 0644)
	_, errEmpty := verifier.VerifyArtifact(ctx, "model_01", "v3.1", emptyPath, "")
	assert.Error(t, errEmpty)
}
