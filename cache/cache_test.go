package cache

import (
	"sync"
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	c := New()
	defer c.Stop()

	c.Set("key1", "value1", time.Minute)
	val, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected key1 to exist")
	}
	if val != "value1" {
		t.Errorf("Get(key1) = %v, want value1", val)
	}
}

func TestCache_Get_Expired(t *testing.T) {
	c := New()
	defer c.Stop()

	c.Set("key1", "value1", 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)

	_, ok := c.Get("key1")
	if ok {
		t.Error("expected expired key to return false")
	}
}

func TestCache_GetString(t *testing.T) {
	c := New()
	defer c.Stop()

	c.Set("str", "hello", time.Minute)
	c.Set("num", 42, time.Minute)

	str, ok := c.GetString("str")
	if !ok || str != "hello" {
		t.Errorf("GetString(str) = %q, %v; want hello, true", str, ok)
	}

	_, ok = c.GetString("num")
	if ok {
		t.Error("GetString(num) should return false for non-string value")
	}

	_, ok = c.GetString("missing")
	if ok {
		t.Error("GetString(missing) should return false")
	}
}

func TestCache_GetInt(t *testing.T) {
	c := New()
	defer c.Stop()

	c.Set("num", 42, time.Minute)
	c.Set("str", "hello", time.Minute)

	num, ok := c.GetInt("num")
	if !ok || num != 42 {
		t.Errorf("GetInt(num) = %d, %v; want 42, true", num, ok)
	}

	_, ok = c.GetInt("str")
	if ok {
		t.Error("GetInt(str) should return false for non-int value")
	}
}

func TestCache_Delete(t *testing.T) {
	c := New()
	defer c.Stop()

	c.Set("key1", "value1", time.Minute)
	c.Delete("key1")

	_, ok := c.Get("key1")
	if ok {
		t.Error("expected key1 to be deleted")
	}
}

func TestCache_Clear(t *testing.T) {
	c := New()
	defer c.Stop()

	c.Set("k1", "v1", time.Minute)
	c.Set("k2", "v2", time.Minute)
	c.Clear()

	if c.Len() != 0 {
		t.Errorf("Len() = %d after Clear, want 0", c.Len())
	}
}

func TestCache_Has(t *testing.T) {
	c := New()
	defer c.Stop()

	c.Set("exists", "val", time.Minute)
	c.Set("expired", "val", 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)

	if !c.Has("exists") {
		t.Error("Has(exists) = false, want true")
	}
	if c.Has("missing") {
		t.Error("Has(missing) = true, want false")
	}
	if c.Has("expired") {
		t.Error("Has(expired) = true, want false for expired key")
	}
}

func TestCache_Len(t *testing.T) {
	c := New()
	defer c.Stop()

	c.Set("k1", "v1", time.Minute)
	c.Set("k2", "v2", 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)

	// Len includes expired items per implementation
	if c.Len() != 2 {
		t.Errorf("Len() = %d, want 2 (includes expired)", c.Len())
	}
}

func TestCache_Keys(t *testing.T) {
	c := New()
	defer c.Stop()

	c.Set("active1", "v1", time.Minute)
	c.Set("active2", "v2", time.Minute)
	c.Set("expired", "v3", 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)

	keys := c.Keys()
	if len(keys) != 2 {
		t.Errorf("Keys() returned %d keys, want 2 (non-expired only)", len(keys))
	}

	keySet := map[string]bool{}
	for _, k := range keys {
		keySet[k] = true
	}
	if !keySet["active1"] || !keySet["active2"] {
		t.Errorf("Keys() = %v, want [active1, active2]", keys)
	}
}

func TestCache_GetOrSet(t *testing.T) {
	c := New()
	defer c.Stop()

	callCount := 0
	fn := func() (interface{}, error) {
		callCount++
		return "computed", nil
	}

	// Miss — should call fn
	val, err := c.GetOrSet("key", time.Minute, fn)
	if err != nil {
		t.Fatalf("GetOrSet error: %v", err)
	}
	if val != "computed" {
		t.Errorf("GetOrSet = %v, want computed", val)
	}
	if callCount != 1 {
		t.Errorf("fn called %d times, want 1", callCount)
	}

	// Hit — should NOT call fn again
	val, err = c.GetOrSet("key", time.Minute, fn)
	if err != nil {
		t.Fatalf("GetOrSet error: %v", err)
	}
	if val != "computed" {
		t.Errorf("GetOrSet = %v, want computed", val)
	}
	if callCount != 1 {
		t.Errorf("fn called %d times on cache hit, want 1", callCount)
	}
}

func TestCache_SetIfNotExists(t *testing.T) {
	c := New()
	defer c.Stop()

	// Set when absent
	ok := c.SetIfNotExists("key", "first", time.Minute)
	if !ok {
		t.Error("SetIfNotExists should return true when key absent")
	}

	// Don't overwrite existing
	ok = c.SetIfNotExists("key", "second", time.Minute)
	if ok {
		t.Error("SetIfNotExists should return false when key exists")
	}

	val, _ := c.Get("key")
	if val != "first" {
		t.Errorf("value = %v, want first (should not be overwritten)", val)
	}

	// Overwrite expired
	c.Set("expiring", "old", 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	ok = c.SetIfNotExists("expiring", "new", time.Minute)
	if !ok {
		t.Error("SetIfNotExists should return true for expired key")
	}
}

func TestCache_Stop(t *testing.T) {
	c := New()
	// Should not panic
	c.Stop()
}

func TestCache_Concurrent(t *testing.T) {
	c := New()
	defer c.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := "key"
			c.Set(key, n, time.Minute)
			c.Get(key)
			c.Has(key)
			c.Keys()
		}(i)
	}
	wg.Wait()
}
