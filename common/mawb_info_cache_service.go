package common

import (
	"context"
	"fmt"
	"time"
)

// MAWBInfoCacheService provides caching functionality for MAWB Info records
type MAWBInfoCacheService struct {
	cache *MAWBInfoCache
}

// NewMAWBInfoCacheService creates a new MAWB Info cache service
func NewMAWBInfoCacheService() *MAWBInfoCacheService {
	// Create memory cache with reasonable limits
	memoryCache := NewMemoryCache(1000)              // Cache up to 1000 MAWB Info records
	memoryCache.StartCleanupRoutine(5 * time.Minute) // Cleanup every 5 minutes

	// Create MAWB Info cache with 20 minute TTL
	mawbInfoCache := NewMAWBInfoCache(memoryCache, 20*time.Minute)

	return &MAWBInfoCacheService{
		cache: mawbInfoCache,
	}
}

// GetMAWBInfo retrieves MAWB Info from cache
func (mics *MAWBInfoCacheService) GetMAWBInfo(ctx context.Context, uuid string) (interface{}, bool) {
	return mics.cache.GetMAWBInfo(ctx, uuid)
}

// SetMAWBInfo stores MAWB Info in cache
func (mics *MAWBInfoCacheService) SetMAWBInfo(ctx context.Context, uuid string, mawbInfo interface{}) error {
	return mics.cache.SetMAWBInfo(ctx, uuid, mawbInfo)
}

// DeleteMAWBInfo removes MAWB Info from cache
func (mics *MAWBInfoCacheService) DeleteMAWBInfo(ctx context.Context, uuid string) error {
	return mics.cache.DeleteMAWBInfo(ctx, uuid)
}

// InvalidateMAWBInfo removes MAWB Info from cache (alias for DeleteMAWBInfo)
func (mics *MAWBInfoCacheService) InvalidateMAWBInfo(ctx context.Context, uuid string) error {
	return mics.DeleteMAWBInfo(ctx, uuid)
}

// GetCacheStats returns cache statistics
func (mics *MAWBInfoCacheService) GetCacheStats() CacheStats {
	return mics.cache.cache.GetStats()
}

// ClearCache clears all cached MAWB Info records
func (mics *MAWBInfoCacheService) ClearCache(ctx context.Context) error {
	return mics.cache.cache.Clear(ctx)
}

// WarmupCache pre-loads frequently accessed MAWB Info records
func (mics *MAWBInfoCacheService) WarmupCache(ctx context.Context, mawbInfoRecords map[string]interface{}) error {
	for uuid, mawbInfo := range mawbInfoRecords {
		if err := mics.SetMAWBInfo(ctx, uuid, mawbInfo); err != nil {
			return fmt.Errorf("failed to warmup cache for MAWB Info %s: %w", uuid, err)
		}
	}
	return nil
}

// GetCacheKey generates a cache key for MAWB Info
func (mics *MAWBInfoCacheService) GetCacheKey(uuid string) string {
	return fmt.Sprintf("mawb_info:%s", uuid)
}

// BatchGetMAWBInfo retrieves multiple MAWB Info records from cache
func (mics *MAWBInfoCacheService) BatchGetMAWBInfo(ctx context.Context, uuids []string) (map[string]interface{}, []string) {
	cached := make(map[string]interface{})
	var missing []string

	for _, uuid := range uuids {
		if mawbInfo, exists := mics.GetMAWBInfo(ctx, uuid); exists {
			cached[uuid] = mawbInfo
		} else {
			missing = append(missing, uuid)
		}
	}

	return cached, missing
}

// BatchSetMAWBInfo stores multiple MAWB Info records in cache
func (mics *MAWBInfoCacheService) BatchSetMAWBInfo(ctx context.Context, mawbInfoRecords map[string]interface{}) error {
	for uuid, mawbInfo := range mawbInfoRecords {
		if err := mics.SetMAWBInfo(ctx, uuid, mawbInfo); err != nil {
			return fmt.Errorf("failed to cache MAWB Info %s: %w", uuid, err)
		}
	}
	return nil
}

// BatchDeleteMAWBInfo removes multiple MAWB Info records from cache
func (mics *MAWBInfoCacheService) BatchDeleteMAWBInfo(ctx context.Context, uuids []string) error {
	for _, uuid := range uuids {
		if err := mics.DeleteMAWBInfo(ctx, uuid); err != nil {
			return fmt.Errorf("failed to delete cached MAWB Info %s: %w", uuid, err)
		}
	}
	return nil
}

// GetCacheHitRatio returns the cache hit ratio as a percentage
func (mics *MAWBInfoCacheService) GetCacheHitRatio() float64 {
	stats := mics.GetCacheStats()
	return stats.HitRatio * 100
}

// GetCacheSize returns the current number of items in cache
func (mics *MAWBInfoCacheService) GetCacheSize() int {
	stats := mics.GetCacheStats()
	return stats.Size
}

// IsHealthy checks if the cache service is functioning properly
func (mics *MAWBInfoCacheService) IsHealthy(ctx context.Context) bool {
	// Test cache functionality with a simple set/get operation
	testKey := "health_check_test"
	testValue := "test_value"

	// Try to set a test value
	if err := mics.cache.cache.Set(ctx, testKey, testValue, 1*time.Second); err != nil {
		return false
	}

	// Try to get the test value
	if value, exists := mics.cache.cache.Get(ctx, testKey); !exists || value != testValue {
		return false
	}

	// Clean up test value
	mics.cache.cache.Delete(ctx, testKey)

	return true
}
