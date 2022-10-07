package cache_test

import (
	"testing"

	"github.com/abiiranathan/gowrap/cache"
)

func TestCache(t *testing.T) {
	c := cache.New[int, int]()

	type pair struct {
		key   int
		value int
	}

	data := []pair{{10, 100}, {30, 50}, {20000, 56000}}
	for _, row := range data {
		c.Put(row.key, row.value)
	}

	// Test Get
	for index, row := range data {
		value, ok := c.Get(row.key)
		if !ok {
			t.Errorf("Get did not retrieve a value")
		}

		if value != data[index].value {
			t.Error("Get returned a wrong value")
		}
	}

	// Test delete
	c.Delete(data[0].key)

	// Make sure there are n
	if _, ok := c.Get(data[0].key); ok {
		t.Errorf("cache key not deleted")
	}

	// test clear
	c.Clear()

	// Make sure there are no values in cache
	for _, row := range data {
		if _, ok := c.Get(row.key); ok {
			t.Errorf("cache key not cleared")
		}
	}
}
