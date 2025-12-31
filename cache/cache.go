// Package cache provides a simple in-memory cache with TTL support
package cache

import (
	"sync"
	"time"
)

// Cache provides a thread-safe in-memory cache with TTL
type Cache struct {
	mu       sync.RWMutex
	items    map[string]cacheItem
	stopChan chan struct{}
}

type cacheItem struct {
	value      interface{}
	expiration time.Time
}

// New creates a new cache instance with automatic cleanup
func New() *Cache {
	c := &Cache{
		items:    make(map[string]cacheItem),
		stopChan: make(chan struct{}),
	}
	go c.cleanup()
	return c
}

// Set stores a value in the cache with the given TTL
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = cacheItem{
		value:      value,
		expiration: time.Now().Add(ttl),
	}
}

// Get retrieves a value from the cache
// Returns the value and a boolean indicating if the key was found and not expired
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(item.expiration) {
		return nil, false
	}

	return item.value, true
}

// GetString retrieves a string value from the cache
func (c *Cache) GetString(key string) (string, bool) {
	val, ok := c.Get(key)
	if !ok {
		return "", false
	}
	str, ok := val.(string)
	return str, ok
}

// GetInt retrieves an int value from the cache
func (c *Cache) GetInt(key string) (int, bool) {
	val, ok := c.Get(key)
	if !ok {
		return 0, false
	}
	i, ok := val.(int)
	return i, ok
}

// Delete removes a key from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Clear removes all items from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]cacheItem)
}

// Has checks if a key exists and is not expired
func (c *Cache) Has(key string) bool {
	_, ok := c.Get(key)
	return ok
}

// Len returns the number of items in the cache (including expired ones)
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Stop stops the cleanup goroutine
func (c *Cache) Stop() {
	close(c.stopChan)
}

// cleanup periodically removes expired items
func (c *Cache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.removeExpired()
		case <-c.stopChan:
			return
		}
	}
}

// removeExpired removes all expired items from the cache
func (c *Cache) removeExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if now.After(item.expiration) {
			delete(c.items, key)
		}
	}
}

// GetOrSet retrieves a value from the cache or sets it using the provided function
func (c *Cache) GetOrSet(key string, ttl time.Duration, fn func() (interface{}, error)) (interface{}, error) {
	if val, ok := c.Get(key); ok {
		return val, nil
	}

	val, err := fn()
	if err != nil {
		return nil, err
	}

	c.Set(key, val, ttl)
	return val, nil
}

// SetIfNotExists sets a value only if the key doesn't exist
// Returns true if the value was set, false if the key already exists
func (c *Cache) SetIfNotExists(key string, value interface{}, ttl time.Duration) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.items[key]
	if exists && time.Now().Before(item.expiration) {
		return false
	}

	c.items[key] = cacheItem{
		value:      value,
		expiration: time.Now().Add(ttl),
	}
	return true
}

// Keys returns all non-expired keys in the cache
func (c *Cache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	keys := make([]string, 0, len(c.items))
	for key, item := range c.items {
		if now.Before(item.expiration) {
			keys = append(keys, key)
		}
	}
	return keys
}
