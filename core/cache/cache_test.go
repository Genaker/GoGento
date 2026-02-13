package cache

import (
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	c := NewCache()

	tests := []struct {
		name     string
		key      string
		value    interface{}
		ttl      int64
		tags     []string
		checkVal func(interface{}) bool
	}{
		{
			name:  "string value with no expiration",
			key:   "test_key",
			value: "test_value",
			ttl:   0,
			tags:  nil,
			checkVal: func(v interface{}) bool {
				return v == "test_value"
			},
		},
		{
			name:  "integer value with tags",
			key:   "test_int",
			value: 42,
			ttl:   0,
			tags:  []string{"numbers", "test"},
			checkVal: func(v interface{}) bool {
				return v == 42
			},
		},
		{
			name:  "struct value",
			key:   "test_struct",
			value: map[string]interface{}{"name": "test", "count": 10},
			ttl:   0,
			tags:  []string{"maps"},
			checkVal: func(v interface{}) bool {
				m, ok := v.(map[string]interface{})
				if !ok {
					return false
				}
				return m["name"] == "test" && m["count"] == 10
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the value
			c.Set(tt.key, tt.value, tt.ttl, tt.tags)

			// Get the value
			got, ok := c.Get(tt.key)
			if !ok {
				t.Errorf("Get() returned ok = false, want true")
				return
			}

			// Check value using custom checker
			if !tt.checkVal(got) {
				t.Errorf("Get() value check failed for %v", got)
			}
		})
	}
}

func TestCache_GetNonExistent(t *testing.T) {
	c := NewCache()

	_, ok := c.Get("non_existent_key")
	if ok {
		t.Errorf("Get() for non-existent key returned ok = true, want false")
	}
}

func TestCache_Delete(t *testing.T) {
	c := NewCache()

	key := "test_delete"
	value := "test_value"

	// Set a value
	c.Set(key, value, 0, nil)

	// Verify it exists
	_, ok := c.Get(key)
	if !ok {
		t.Fatalf("Get() returned ok = false after Set(), want true")
	}

	// Delete the value
	c.Delete(key)

	// Verify it's deleted
	_, ok = c.Get(key)
	if ok {
		t.Errorf("Get() after Delete() returned ok = true, want false")
	}
}

func TestCache_Expiration(t *testing.T) {
	c := NewCache()

	key := "test_expire"
	value := "test_value"
	ttl := int64(1) // 1 second

	// Set a value with short TTL
	c.Set(key, value, ttl, nil)

	// Immediately get the value - should exist
	got, ok := c.Get(key)
	if !ok {
		t.Fatalf("Get() immediately after Set() returned ok = false, want true")
	}
	if got != value {
		t.Errorf("Get() = %v, want %v", got, value)
	}

	// Wait for expiration
	time.Sleep(1100 * time.Millisecond)

	// Try to get expired value - should not exist
	_, ok = c.Get(key)
	if ok {
		t.Errorf("Get() after expiration returned ok = true, want false")
	}
}

func TestGetInstance(t *testing.T) {
	// Get instance twice
	inst1 := GetInstance()
	inst2 := GetInstance()

	// Should be the same instance (singleton)
	if inst1 != inst2 {
		t.Errorf("GetInstance() should return the same instance (singleton pattern)")
	}

	// Set a value using inst1
	inst1.Set("singleton_test_unique", "test_value", 0, nil)

	// Get the value using inst2
	val, ok := inst2.Get("singleton_test_unique")
	if !ok {
		t.Errorf("Value set on inst1 should be accessible from inst2")
	}
	if val != "test_value" {
		t.Errorf("Got value %v, want 'test_value'", val)
	}
}
