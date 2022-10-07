package cache

import (
	"github.com/abiiranathan/gowrap/async"
)

// A simple interface abstraction on top of async.AsyncMap
type Cache[K comparable, V any] interface {
	// Store value in cache
	Put(key K, value V)

	// Retrieve value from cache.
	// if the value does not exist, exists is false
	Get(key K) (value V, exists bool)

	// Delete value from cache
	Delete(key K)

	// Clear the cache
	Clear()
}

type storage[K comparable, V any] struct {
	data *async.AsyncMap[K, V]
}

func New[K comparable, V any]() Cache[K, V] {
	return &storage[K, V]{
		data: async.NewAsyncMap[K, V](),
	}
}

func (s *storage[K, V]) Put(key K, value V) {
	s.data.Set(key, value)
}

func (s *storage[K, V]) Get(key K) (V, bool) {
	return s.data.Get(key)
}

func (s *storage[K, V]) Delete(key K) {
	s.data.Delete(key)
}

func (s *storage[K, V]) Clear() {
	s.data.Clear()
}
