package api

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestTTLCache(t *testing.T) {
	t.Parallel()
	var cache TTLCache
	if value, ok := cache.Get("missing"); ok || value != nil {
		t.Fatalf("missing key = %v, %v", value, ok)
	}
	cache.Set("live", "original", time.Hour)
	cache.Set("expired", "old", -time.Second)
	if value, ok := cache.Get("expired"); ok || value != nil {
		t.Fatalf("expired key = %v, %v", value, ok)
	}
	cache.Set("live", "updated", time.Hour)
	if value, ok := cache.Get("live"); !ok || value != "updated" {
		t.Fatalf("updated key = %v, %v", value, ok)
	}
	cache.Set("other", "preserved", time.Hour)
	cache.Invalidate("live", "missing")
	if _, ok := cache.Get("live"); ok {
		t.Fatal("invalidated value retained")
	}
	if value, ok := cache.Get("other"); !ok || value != "preserved" {
		t.Fatalf("invalidation changed unrelated key: %v, %v", value, ok)
	}
}

func TestTTLCacheConcurrentAccess(t *testing.T) {
	t.Parallel()
	var cache TTLCache
	var wg sync.WaitGroup
	for worker := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			key := fmt.Sprintf("worker-%d", worker)
			for i := range 100 {
				cache.Set(key, i, time.Hour)
				if got, ok := cache.Get(key); !ok || got != i {
					t.Errorf("Get(%q) = %v, %v, want %d", key, got, ok, i)
				}
				cache.Invalidate(key)
				cache.Set("shared", i, time.Hour)
				cache.Get("shared")
				cache.Invalidate("shared")
			}
		}()
	}
	wg.Wait()
}
