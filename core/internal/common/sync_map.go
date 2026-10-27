package common

import "sync"

// SyncMap is a type-safe wrapper around sync.Map: Load and LoadOrStore keep sync.Map's return
// shape but take and return K/V, so callers never type-assert. The zero value is ready to use.
//
// sync.Map is embedded, so its other methods (Delete, Store, Range, ...) are still promoted and
// stay untyped; wrap them here as they are needed.
type SyncMap[K comparable, V any] struct {
	sync.Map
}

// Load returns the value stored for key, or V's zero value and false if there is none.
func (m *SyncMap[K, V]) Load(key K) (V, bool) {
	value, ok := m.Map.Load(key)
	if !ok {
		var zero V
		return zero, false
	}
	typed, _ := value.(V) // comma-ok: a stored nil interface value yields V's zero value, not a panic
	return typed, true
}

// LoadOrStore returns the existing value for key if present (loaded=true); otherwise it stores
// and returns value (loaded=false).
func (m *SyncMap[K, V]) LoadOrStore(key K, value V) (V, bool) {
	actual, loaded := m.Map.LoadOrStore(key, value)
	typed, _ := actual.(V)
	return typed, loaded
}
