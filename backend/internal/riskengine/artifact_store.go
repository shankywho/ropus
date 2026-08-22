package riskengine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// ArtifactStore defines the storage abstraction for candidate and production model artifacts.
type ArtifactStore interface {
	StoreArtifact(ctx context.Context, modelID, filename string, r io.Reader) (storageURI string, checksum string, err error)
	GetArtifact(ctx context.Context, storageURI string) (io.ReadCloser, error)
	ArtifactExists(ctx context.Context, storageURI string) bool
	GetChecksum(ctx context.Context, storageURI string) (string, error)
	GetBaseDir() string
}

// LocalFilesystemArtifactStore stores model artifacts immutably on the local filesystem.
type LocalFilesystemArtifactStore struct {
	baseDir string
	mu      sync.RWMutex
}

// NewLocalFilesystemArtifactStore initializes an immutable local filesystem artifact store.
func NewLocalFilesystemArtifactStore(baseDir string) (*LocalFilesystemArtifactStore, error) {
	if baseDir == "" {
		baseDir = "./ml-service/model/candidates"
	}
	absDir, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve artifact store path '%s': %w", baseDir, err)
	}

	if err := os.MkdirAll(absDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create artifact store directory '%s': %w", absDir, err)
	}

	return &LocalFilesystemArtifactStore{
		baseDir: absDir,
	}, nil
}

// GetBaseDir returns the base root directory for artifact storage.
func (s *LocalFilesystemArtifactStore) GetBaseDir() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.baseDir
}

// StoreArtifact writes an artifact stream to the store and enforces immutability.
func (s *LocalFilesystemArtifactStore) StoreArtifact(ctx context.Context, modelID, filename string, r io.Reader) (string, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if modelID == "" || filename == "" {
		return "", "", fmt.Errorf("modelID and filename cannot be empty")
	}

	// Sanitize filename to prevent directory traversal
	cleanFilename := filepath.Base(filename)
	modelDir := filepath.Join(s.baseDir, modelID)
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create model artifact directory: %w", err)
	}

	targetPath := filepath.Join(modelDir, cleanFilename)

	// IMMUTABILITY GUARANTEE: Never overwrite an existing model artifact
	if _, err := os.Stat(targetPath); err == nil {
		return "", "", fmt.Errorf("artifact immutability violation: file '%s' already exists for model '%s'", cleanFilename, modelID)
	}

	tmpPath := targetPath + ".tmp"
	outFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return "", "", fmt.Errorf("failed to create temporary artifact file: %w", err)
	}

	hasher := sha256.New()
	multiWriter := io.MultiWriter(outFile, hasher)

	if _, err := io.Copy(multiWriter, r); err != nil {
		outFile.Close()
		_ = os.Remove(tmpPath)
		return "", "", fmt.Errorf("failed to write artifact data: %w", err)
	}
	outFile.Close()

	if err := os.Rename(tmpPath, targetPath); err != nil {
		_ = os.Remove(tmpPath)
		return "", "", fmt.Errorf("failed to finalize artifact file: %w", err)
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))
	storageURI := fmt.Sprintf("file://%s", targetPath)

	return storageURI, checksum, nil
}

// GetArtifact opens an artifact file for reading.
func (s *LocalFilesystemArtifactStore) GetArtifact(ctx context.Context, storageURI string) (io.ReadCloser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filePath := resolveFilePath(storageURI)
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open artifact at '%s': %w", storageURI, err)
	}
	return f, nil
}

// ArtifactExists checks if an artifact is present and non-empty.
func (s *LocalFilesystemArtifactStore) ArtifactExists(ctx context.Context, storageURI string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filePath := resolveFilePath(storageURI)
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() || info.Size() == 0 {
		return false
	}
	return true
}

// GetChecksum calculates or verifies the SHA-256 checksum of an existing artifact.
func (s *LocalFilesystemArtifactStore) GetChecksum(ctx context.Context, storageURI string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filePath := resolveFilePath(storageURI)
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open artifact for checksum calculation: %w", err)
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func resolveFilePath(storageURI string) string {
	if len(storageURI) > 7 && storageURI[:7] == "file://" {
		return storageURI[7:]
	}
	return storageURI
}
