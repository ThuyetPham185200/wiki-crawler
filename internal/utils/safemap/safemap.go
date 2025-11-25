package safemap

import (
	"sync"
)

// SafeMap is a concurrent-safe map using a RWMutex.
// K must be comparable (map key constraint), V can be any type.
type SafeMap[K comparable, V any] struct {
	mu sync.RWMutex
	m  map[K]V
}

// New creates and returns a new SafeMap.
func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
	return &SafeMap[K, V]{
		m: make(map[K]V),
	}
}

// Set sets key -> value.
func (s *SafeMap[K, V]) Set(key K, value V) {
	s.mu.Lock()
	s.m[key] = value
	s.mu.Unlock()
}

// Get returns value and true if key exists, otherwise zero value and false.
func (s *SafeMap[K, V]) Get(key K) (V, bool) {
	s.mu.RLock()
	v, ok := s.m[key]
	s.mu.RUnlock()
	return v, ok
}

// Delete removes the key from the map.
func (s *SafeMap[K, V]) Delete(key K) {
	s.mu.Lock()
	delete(s.m, key)
	s.mu.Unlock()
}

// Len returns the number of elements in the map.
func (s *SafeMap[K, V]) Len() int {
	s.mu.RLock()
	n := len(s.m)
	s.mu.RUnlock()
	return n
}

// Keys returns a slice of keys (snapshot). Order is unspecified.
func (s *SafeMap[K, V]) Keys() []K {
	s.mu.RLock()
	keys := make([]K, 0, len(s.m))
	for k := range s.m {
		keys = append(keys, k)
	}
	s.mu.RUnlock()
	return keys
}

// Values returns a slice of values (snapshot). Order is unspecified.
func (s *SafeMap[K, V]) Values() []V {
	s.mu.RLock()
	values := make([]V, 0, len(s.m))
	for _, v := range s.m {
		values = append(values, v)
	}
	s.mu.RUnlock()
	return values
}

// LoadOrStore returns existing value if present.
// If not present, it stores and returns the given value.
// The loaded boolean is true if the value was already present.
func (s *SafeMap[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v, ok := s.m[key]; ok {
		return v, true
	}
	s.m[key] = value
	return value, false
}

// GetOrDefault returns value if present otherwise defaultVal.
func (s *SafeMap[K, V]) GetOrDefault(key K, defaultVal V) V {
	s.mu.RLock()
	v, ok := s.m[key]
	s.mu.RUnlock()
	if !ok {
		return defaultVal
	}
	return v
}

// Range calls fn for each key/value pair. If fn returns false, iteration stops.
// It holds the read lock only while iterating; fn should not call back into this map.
func (s *SafeMap[K, V]) Range(fn func(k K, v V) bool) {
	s.mu.RLock()
	// create snapshot of entries to avoid holding lock while calling user function
	snapshot := make([]struct {
		k K
		v V
	}, 0, len(s.m))
	for k, v := range s.m {
		snapshot = append(snapshot, struct {
			k K
			v V
		}{k: k, v: v})
	}
	s.mu.RUnlock()

	for _, e := range snapshot {
		if !fn(e.k, e.v) {
			return
		}
	}
}

// Clear removes all entries.
func (s *SafeMap[K, V]) Clear() {
	s.mu.Lock()
	s.m = make(map[K]V)
	s.mu.Unlock()
}

// ⚡️ ContainsKey checks if a key exists
func (s *SafeMap[K, V]) ContainsKey(key K) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.m[key]
	return ok
}
