package store

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// OnlineStore provides sub-millisecond real-time feature lookup for live inference.
type OnlineStore interface {
	GetOnlineFeatures(ctx context.Context, entityKey string, featureNames []string) (map[string]interface{}, error)
	PutOnlineFeatures(ctx context.Context, entityKey string, features map[string]interface{}, ttl time.Duration) error
}

// OfflineStore provides historical batch feature retrieval and point-in-time snapshotting.
type OfflineStore interface {
	SaveSnapshot(ctx context.Context, snapshot *FeatureSnapshot) error
	GetSnapshot(ctx context.Context, snapshotID string) (*FeatureSnapshot, error)
}

// FeatureStore unifies online feature serving and offline dataset generation.
type FeatureStore struct {
	registry *FeatureRegistry
	online   OnlineStore
	offline  OfflineStore
}

// NewFeatureStore initializes the unified feature store.
func NewFeatureStore(reg *FeatureRegistry, online OnlineStore, offline OfflineStore) *FeatureStore {
	if reg == nil {
		reg = NewFeatureRegistry()
	}
	if online == nil {
		online = NewInMemoryOnlineStore()
	}
	if offline == nil {
		offline = NewInMemoryOfflineStore()
	}
	return &FeatureStore{
		registry: reg,
		online:   online,
		offline:  offline,
	}
}

func (s *FeatureStore) Registry() *FeatureRegistry {
	return s.registry
}

func (s *FeatureStore) Online() OnlineStore {
	return s.online
}

func (s *FeatureStore) Offline() OfflineStore {
	return s.offline
}

// ---------------------------------------------------------------------------
// In-Memory Implementations (Thread-Safe Baseline)
// ---------------------------------------------------------------------------
type InMemoryOnlineStore struct {
	mu   sync.RWMutex
	data map[string]map[string]interface{}
}

func NewInMemoryOnlineStore() *InMemoryOnlineStore {
	return &InMemoryOnlineStore{
		data: make(map[string]map[string]interface{}),
	}
}

func (s *InMemoryOnlineStore) GetOnlineFeatures(ctx context.Context, entityKey string, featureNames []string) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entityData, exists := s.data[entityKey]
	if !exists {
		return make(map[string]interface{}), nil
	}

	result := make(map[string]interface{}, len(featureNames))
	for _, name := range featureNames {
		if val, found := entityData[name]; found {
			result[name] = val
		}
	}
	return result, nil
}

func (s *InMemoryOnlineStore) PutOnlineFeatures(ctx context.Context, entityKey string, features map[string]interface{}, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.data[entityKey]; !exists {
		s.data[entityKey] = make(map[string]interface{})
	}
	for k, v := range features {
		s.data[entityKey][k] = v
	}
	return nil
}

type InMemoryOfflineStore struct {
	mu        sync.RWMutex
	snapshots map[string]*FeatureSnapshot
}

func NewInMemoryOfflineStore() *InMemoryOfflineStore {
	return &InMemoryOfflineStore{
		snapshots: make(map[string]*FeatureSnapshot),
	}
}

func (s *InMemoryOfflineStore) SaveSnapshot(ctx context.Context, snapshot *FeatureSnapshot) error {
	if snapshot == nil {
		return fmt.Errorf("snapshot is nil")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshots[snapshot.SnapshotID] = snapshot
	return nil
}

func (s *InMemoryOfflineStore) GetSnapshot(ctx context.Context, snapshotID string) (*FeatureSnapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snap, exists := s.snapshots[snapshotID]
	if !exists {
		return nil, fmt.Errorf("snapshot '%s' not found", snapshotID)
	}
	return snap, nil
}
