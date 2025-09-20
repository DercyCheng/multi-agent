package temporal

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MockCacheManager provides a mock implementation of CacheManager for testing
type MockCacheManager struct {
	logger *zap.Logger
	cache  map[string]*CacheEntry
	mu     sync.RWMutex
}

// CacheEntry represents a cached entry
type CacheEntry struct {
	Value     interface{}
	ExpiresAt time.Time
}

// NewMockCacheManager creates a new mock cache manager
func NewMockCacheManager(logger *zap.Logger) *MockCacheManager {
	return &MockCacheManager{
		logger: logger,
		cache:  make(map[string]*CacheEntry),
	}
}

// Get retrieves a value from cache
func (m *MockCacheManager) Get(ctx context.Context, key string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.cache[key]
	if !exists {
		return nil, fmt.Errorf("cache miss: key not found")
	}

	if time.Now().After(entry.ExpiresAt) {
		delete(m.cache, key)
		return nil, fmt.Errorf("cache miss: key expired")
	}

	m.logger.Debug("Cache hit", zap.String("key", key))
	return entry.Value, nil
}

// Set stores a value in cache
func (m *MockCacheManager) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cache[key] = &CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}

	m.logger.Debug("Cache set", zap.String("key", key), zap.Duration("ttl", ttl))
	return nil
}

// GetPattern retrieves values matching a pattern
func (m *MockCacheManager) GetPattern(ctx context.Context, pattern string) ([]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []interface{}
	now := time.Now()

	for key, entry := range m.cache {
		if now.After(entry.ExpiresAt) {
			continue
		}
		
		// Simple pattern matching (contains)
		if contains(key, pattern) {
			results = append(results, entry.Value)
		}
	}

	m.logger.Debug("Pattern search", zap.String("pattern", pattern), zap.Int("results", len(results)))
	return results, nil
}

// SetPattern stores a value with pattern-based key
func (m *MockCacheManager) SetPattern(ctx context.Context, pattern string, value interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("pattern:%s:%d", pattern, time.Now().UnixNano())
	return m.Set(ctx, key, value, ttl)
}

// CleanupExpired removes expired entries
func (m *MockCacheManager) CleanupExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	expiredCount := 0

	for key, entry := range m.cache {
		if now.After(entry.ExpiresAt) {
			delete(m.cache, key)
			expiredCount++
		}
	}

	if expiredCount > 0 {
		m.logger.Info("Cleaned up expired cache entries", zap.Int("count", expiredCount))
	}
}

// Helper function for simple pattern matching
func contains(s, pattern string) bool {
	return len(pattern) == 0 || (len(s) >= len(pattern) && s[:len(pattern)] == pattern)
}