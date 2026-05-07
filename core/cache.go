package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CacheManager manages cached API data
type CacheManager struct {
	cacheDir string
	cache    map[string]*CacheEntry
	mu       sync.RWMutex
}

// CacheEntry represents a cached data entry
type CacheEntry struct {
	Data      *UsageData `json:"data"`
	CachedAt  time.Time  `json:"cachedAt"`
	ExpiresAt time.Time  `json:"expiresAt"`
}

// NewCacheManager creates a new cache manager
func NewCacheManager() (*CacheManager, error) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, err
	}

	cm := &CacheManager{
		cacheDir: cacheDir,
		cache:    make(map[string]*CacheEntry),
	}

	// Load existing cache from disk
	if err := cm.loadAll(); err != nil {
		// Ignore errors loading cache, start fresh
	}

	return cm, nil
}

func getCacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".cloudpulse", "cache"), nil
}

func (cm *CacheManager) getCacheKey(platform, accountName string) string {
	return platform + "_" + accountName
}

func (cm *CacheManager) loadAll() error {
	entries, err := os.ReadDir(cm.cacheDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(cm.cacheDir, entry.Name()))
		if err != nil {
			continue
		}

		var cacheEntry CacheEntry
		if err := json.Unmarshal(data, &cacheEntry); err != nil {
			continue
		}

		key := entry.Name()[:len(entry.Name())-5] // remove .json
		cm.cache[key] = &cacheEntry
	}

	return nil
}

func (cm *CacheManager) saveToDisk(key string, entry *CacheEntry) error {
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(cm.cacheDir, key+".json"), data, 0644)
}

// Get returns cached data for a platform/account if not expired
func (cm *CacheManager) Get(platform, accountName string) *UsageData {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	key := cm.getCacheKey(platform, accountName)
	entry, ok := cm.cache[key]
	if !ok {
		return nil
	}

	if time.Now().After(entry.ExpiresAt) {
		return nil
	}

	return entry.Data
}

// Set stores data in the cache with a TTL
func (cm *CacheManager) Set(platform, accountName string, data *UsageData, ttl time.Duration) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	key := cm.getCacheKey(platform, accountName)
	entry := &CacheEntry{
		Data:      data,
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(ttl),
	}

	cm.cache[key] = entry
	return cm.saveToDisk(key, entry)
}

// Delete removes cached data for a platform/account
func (cm *CacheManager) Delete(platform, accountName string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	key := cm.getCacheKey(platform, accountName)
	delete(cm.cache, key)

	cacheFile := filepath.Join(cm.cacheDir, key+".json")
	if err := os.Remove(cacheFile); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// Clear removes all cached data
func (cm *CacheManager) Clear() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.cache = make(map[string]*CacheEntry)

	entries, err := os.ReadDir(cm.cacheDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			os.Remove(filepath.Join(cm.cacheDir, entry.Name()))
		}
	}

	return nil
}
