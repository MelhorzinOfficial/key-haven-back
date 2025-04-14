package cache

import (
	"key-haven-back/internal/service/dto"
	"sync"
	"time"
)

// UserCache implementa um cache simples para armazenar informações de usuário
// com expiração automática para garantir que os dados não fiquem desatualizados.
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

// NewUserCache cria uma nova instância de UserCache com o TTL especificado
func NewUserCache(ttl time.Duration, cleanupInterval time.Duration) *UserCache {
	cache := &UserCache{
		cache:           make(map[string]*cacheEntry),
		ttl:             ttl,
		cleanupInterval: cleanupInterval,
	}

	// Inicia a limpeza periódica do cache
	go cache.startCleanupTimer()

	return cache
}

// Set armazena um usuário no cache com expiração
func (c *UserCache) Set(userID string, user *dto.UserResponse) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache[userID] = &cacheEntry{
		user:      user,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// Get recupera um usuário do cache se ele existir e não tiver expirado
func (c *UserCache) Get(userID string) (*dto.UserResponse, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.cache[userID]
	if !exists {
		return nil, false
	}

	// Verifica se a entrada expirou
	if time.Now().After(entry.expiresAt) {
		// Expirou, mas vamos remover em outra goroutine para não segurar o lock
		go c.Delete(userID)
		return nil, false
	}

	return entry.user, true
}

// Delete remove um usuário do cache
func (c *UserCache) Delete(userID string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.cache, userID)
}

// Limpa entradas expiradas do cache
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

// Inicia um timer para limpar o cache periodicamente
func (c *UserCache) startCleanupTimer() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		<-ticker.C
		c.cleanup()
	}
}
