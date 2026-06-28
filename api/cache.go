// api/cache.go
// 프로세스 내 TTL 캐시 (TUI 전용)

package api

import (
	"sync"
	"time"
)

type cacheEntry struct {
	value     interface{}
	expiresAt time.Time
}

// TTLCache는 sync.Map 기반 단순 TTL 캐시
type TTLCache struct {
	mu sync.Map
}

// Set은 키에 값을 저장하고 TTL을 설정
func (c *TTLCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Store(key, cacheEntry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	})
}

// Get은 키에 해당하는 값을 반환. 만료 시 (nil, false) 반환
func (c *TTLCache) Get(key string) (interface{}, bool) {
	raw, ok := c.mu.Load(key)
	if !ok {
		return nil, false
	}
	entry := raw.(cacheEntry)
	if time.Now().After(entry.expiresAt) {
		c.mu.Delete(key)
		return nil, false
	}
	return entry.value, true
}

// Invalidate는 지정 키들을 즉시 무효화
func (c *TTLCache) Invalidate(keys ...string) {
	for _, k := range keys {
		c.mu.Delete(k)
	}
}

// 캐시 키 상수
const (
	CacheKeyUser     = "user"
	CacheKeySections = "sections"
	// accounts 캐시 키는 "accounts:{section_id}" 형태로 사용
)
