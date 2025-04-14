package cache

import (
	"key-haven-back/internal/service/dto"
	"sync"
	"time"
)

type UserCache struct {
	cache           map[string]*cacheEntry
	mutex           sync.RWMutex
	ttl             time.Duration
	cleanupInterval time.Duration
}

type cacheEntry struct {
	user      *dto.UserResponse
	expiresAt time.Time
}

func NewUserCache(ttl time.Duration, cleanupInterval time.Duration) *UserCache {
	cache := &UserCache{
		cache:           make(map[string]*cacheEntry),
		ttl:             ttl,
		cleanupInterval: cleanupInterval,
	}

	go cache.startCleanupTimer()

	return cache
}

func (c *UserCache) Set(userID string, user *dto.UserResponse) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache[userID] = &cacheEntry{
		user:      user,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *UserCache) Get(userID string) (*dto.UserResponse, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.cache[userID]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		go c.Delete(userID)
		return nil, false
	}

	return entry.user, true
}

func (c *UserCache) Delete(userID string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.cache, userID)
}

func (c *UserCache) cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for id, entry := range c.cache {
		if now.After(entry.expiresAt) {
			delete(c.cache, id)
		}
	}
}

func (c *UserCache) startCleanupTimer() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		<-ticker.C
		c.cleanup()
	}
}
