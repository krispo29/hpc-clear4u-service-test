package common

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// CacheItem represents a cached item with expiration
type CacheItem struct {
	Value     interface{} `json:"value"`
	ExpiresAt time.Time   `json:"expires_at"`
	CreatedAt time.Time   `json:"created_at"`
}

// IsExpired checks if the cache item has expired
func (ci *CacheItem) IsExpired() bool {
	return time.Now().After(ci.ExpiresAt)
}

// Cache interface defines the contract for caching implementations
type Cache interface {
	Get(ctx context.Context, key string) (interface{}, bool)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	GetStats() CacheStats
}

// CacheStats represents cache statistics
type CacheStats struct {
	Hits      int64   `json:"hits"`
	Misses    int64   `json:"misses"`
	Sets      int64   `json:"sets"`
	Deletes   int64   `json:"deletes"`
	Evictions int64   `json:"evictions"`
	Size      int     `json:"size"`
	MaxSize   int     `json:"max_size"`
	HitRatio  float64 `json:"hit_ratio"`
}

// MemoryCache implements an in-memory cache with LRU eviction
type MemoryCache struct {
	items   map[string]*CacheItem
	mutex   sync.RWMutex
	maxSize int
	stats   CacheStats
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(maxSize int) *MemoryCache {
	return &MemoryCache{
		items:   make(map[string]*CacheItem),
		maxSize: maxSize,
		stats:   CacheStats{MaxSize: maxSize},
	}
}

// Get retrieves a value from the cache
func (mc *MemoryCache) Get(ctx context.Context, key string) (interface{}, bool) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	item, exists := mc.items[key]
	if !exists {
		mc.stats.Misses++
		mc.updateHitRatio()
		return nil, false
	}

	if item.IsExpired() {
		// Remove expired item
		delete(mc.items, key)
		mc.stats.Misses++
		mc.stats.Evictions++
		mc.updateHitRatio()
		return nil, false
	}

	mc.stats.Hits++
	mc.updateHitRatio()
	return item.Value, true
}

// Set stores a value in the cache with TTL
func (mc *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	// Check if we need to evict items
	if len(mc.items) >= mc.maxSize {
		mc.evictOldest()
	}

	mc.items[key] = &CacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
		CreatedAt: time.Now(),
	}

	mc.stats.Sets++
	mc.stats.Size = len(mc.items)
	return nil
}

// Delete removes a value from the cache
func (mc *MemoryCache) Delete(ctx context.Context, key string) error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if _, exists := mc.items[key]; exists {
		delete(mc.items, key)
		mc.stats.Deletes++
		mc.stats.Size = len(mc.items)
	}

	return nil
}

// Clear removes all items from the cache
func (mc *MemoryCache) Clear(ctx context.Context) error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	mc.items = make(map[string]*CacheItem)
	mc.stats.Size = 0
	return nil
}

// GetStats returns cache statistics
func (mc *MemoryCache) GetStats() CacheStats {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	stats := mc.stats
	stats.Size = len(mc.items)
	return stats
}

// evictOldest removes the oldest item from the cache
func (mc *MemoryCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range mc.items {
		if oldestKey == "" || item.CreatedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.CreatedAt
		}
	}

	if oldestKey != "" {
		delete(mc.items, oldestKey)
		mc.stats.Evictions++
	}
}

// updateHitRatio calculates the current hit ratio
func (mc *MemoryCache) updateHitRatio() {
	total := mc.stats.Hits + mc.stats.Misses
	if total > 0 {
		mc.stats.HitRatio = float64(mc.stats.Hits) / float64(total)
	}
}

// cleanupExpired removes expired items from the cache
func (mc *MemoryCache) cleanupExpired() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	for key, item := range mc.items {
		if item.IsExpired() {
			delete(mc.items, key)
			mc.stats.Evictions++
		}
	}
	mc.stats.Size = len(mc.items)
}

// StartCleanupRoutine starts a background routine to clean up expired items
func (mc *MemoryCache) StartCleanupRoutine(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			mc.cleanupExpired()
		}
	}()
}

// MAWBInfoCache provides specialized caching for MAWB Info records
type MAWBInfoCache struct {
	cache Cache
	ttl   time.Duration
}

// NewMAWBInfoCache creates a new MAWB Info cache
func NewMAWBInfoCache(cache Cache, ttl time.Duration) *MAWBInfoCache {
	return &MAWBInfoCache{
		cache: cache,
		ttl:   ttl,
	}
}

// GetMAWBInfo retrieves MAWB Info from cache
func (mic *MAWBInfoCache) GetMAWBInfo(ctx context.Context, uuid string) (interface{}, bool) {
	key := fmt.Sprintf("mawb_info:%s", uuid)
	return mic.cache.Get(ctx, key)
}

// SetMAWBInfo stores MAWB Info in cache
func (mic *MAWBInfoCache) SetMAWBInfo(ctx context.Context, uuid string, mawbInfo interface{}) error {
	key := fmt.Sprintf("mawb_info:%s", uuid)
	return mic.cache.Set(ctx, key, mawbInfo, mic.ttl)
}

// DeleteMAWBInfo removes MAWB Info from cache
func (mic *MAWBInfoCache) DeleteMAWBInfo(ctx context.Context, uuid string) error {
	key := fmt.Sprintf("mawb_info:%s", uuid)
	return mic.cache.Delete(ctx, key)
}

// CalculationCache provides caching for complex calculations
type CalculationCache struct {
	cache Cache
	ttl   time.Duration
}

// NewCalculationCache creates a new calculation cache
func NewCalculationCache(cache Cache, ttl time.Duration) *CalculationCache {
	return &CalculationCache{
		cache: cache,
		ttl:   ttl,
	}
}

// CalculationKey represents a key for calculation caching
type CalculationKey struct {
	Type       string      `json:"type"`
	Parameters interface{} `json:"parameters"`
}

// GetCalculationKey generates a cache key for calculations
func (ck *CalculationKey) GetKey() (string, error) {
	data, err := json.Marshal(ck)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("calc:%x", data), nil
}

// GetVolumetricWeight retrieves cached volumetric weight calculation
func (cc *CalculationCache) GetVolumetricWeight(ctx context.Context, dimensions interface{}) (float64, bool) {
	key := &CalculationKey{
		Type:       "volumetric_weight",
		Parameters: dimensions,
	}

	keyStr, err := key.GetKey()
	if err != nil {
		return 0, false
	}

	value, exists := cc.cache.Get(ctx, keyStr)
	if !exists {
		return 0, false
	}

	if weight, ok := value.(float64); ok {
		return weight, true
	}

	return 0, false
}

// SetVolumetricWeight stores calculated volumetric weight in cache
func (cc *CalculationCache) SetVolumetricWeight(ctx context.Context, dimensions interface{}, weight float64) error {
	key := &CalculationKey{
		Type:       "volumetric_weight",
		Parameters: dimensions,
	}

	keyStr, err := key.GetKey()
	if err != nil {
		return err
	}

	return cc.cache.Set(ctx, keyStr, weight, cc.ttl)
}

// GetChargeableWeight retrieves cached chargeable weight calculation
func (cc *CalculationCache) GetChargeableWeight(ctx context.Context, parameters interface{}) (float64, bool) {
	key := &CalculationKey{
		Type:       "chargeable_weight",
		Parameters: parameters,
	}

	keyStr, err := key.GetKey()
	if err != nil {
		return 0, false
	}

	value, exists := cc.cache.Get(ctx, keyStr)
	if !exists {
		return 0, false
	}

	if weight, ok := value.(float64); ok {
		return weight, true
	}

	return 0, false
}

// SetChargeableWeight stores calculated chargeable weight in cache
func (cc *CalculationCache) SetChargeableWeight(ctx context.Context, parameters interface{}, weight float64) error {
	key := &CalculationKey{
		Type:       "chargeable_weight",
		Parameters: parameters,
	}

	keyStr, err := key.GetKey()
	if err != nil {
		return err
	}

	return cc.cache.Set(ctx, keyStr, weight, cc.ttl)
}

// GetFinancialTotals retrieves cached financial totals calculation
func (cc *CalculationCache) GetFinancialTotals(ctx context.Context, charges interface{}) (float64, bool) {
	key := &CalculationKey{
		Type:       "financial_totals",
		Parameters: charges,
	}

	keyStr, err := key.GetKey()
	if err != nil {
		return 0, false
	}

	value, exists := cc.cache.Get(ctx, keyStr)
	if !exists {
		return 0, false
	}

	if total, ok := value.(float64); ok {
		return total, true
	}

	return 0, false
}

// SetFinancialTotals stores calculated financial totals in cache
func (cc *CalculationCache) SetFinancialTotals(ctx context.Context, charges interface{}, total float64) error {
	key := &CalculationKey{
		Type:       "financial_totals",
		Parameters: charges,
	}

	keyStr, err := key.GetKey()
	if err != nil {
		return err
	}

	return cc.cache.Set(ctx, keyStr, total, cc.ttl)
}

// PDFTemplateCache provides caching for PDF templates
type PDFTemplateCache struct {
	cache Cache
	ttl   time.Duration
}

// NewPDFTemplateCache creates a new PDF template cache
func NewPDFTemplateCache(cache Cache, ttl time.Duration) *PDFTemplateCache {
	return &PDFTemplateCache{
		cache: cache,
		ttl:   ttl,
	}
}

// GetTemplate retrieves a cached PDF template
func (ptc *PDFTemplateCache) GetTemplate(ctx context.Context, templateName string) ([]byte, bool) {
	key := fmt.Sprintf("pdf_template:%s", templateName)
	value, exists := ptc.cache.Get(ctx, key)
	if !exists {
		return nil, false
	}

	if template, ok := value.([]byte); ok {
		return template, true
	}

	return nil, false
}

// SetTemplate stores a PDF template in cache
func (ptc *PDFTemplateCache) SetTemplate(ctx context.Context, templateName string, template []byte) error {
	key := fmt.Sprintf("pdf_template:%s", templateName)
	return ptc.cache.Set(ctx, key, template, ptc.ttl)
}

// GetFont retrieves a cached font
func (ptc *PDFTemplateCache) GetFont(ctx context.Context, fontName string) ([]byte, bool) {
	key := fmt.Sprintf("pdf_font:%s", fontName)
	value, exists := ptc.cache.Get(ctx, key)
	if !exists {
		return nil, false
	}

	if font, ok := value.([]byte); ok {
		return font, true
	}

	return nil, false
}

// SetFont stores a font in cache
func (ptc *PDFTemplateCache) SetFont(ctx context.Context, fontName string, font []byte) error {
	key := fmt.Sprintf("pdf_font:%s", fontName)
	return ptc.cache.Set(ctx, key, font, ptc.ttl)
}
