package cache

import (
	"testing"
	"time"
)

func TestCacheSetGet(t *testing.T) {
	c := NewCache()
	c.Set("key1", "value1", 1*time.Second)

	val, found := c.Get("key1")
	if !found {
		t.Error("expected key1 to be found")
	}
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}
}

func TestCacheExpiry(t *testing.T) {
	c := NewCache()
	c.Set("key1", "value1", 100*time.Millisecond)

	time.Sleep(150 * time.Millisecond)

	_, found := c.Get("key1")
	if found {
		t.Error("expected key1 to be expired")
	}
}
