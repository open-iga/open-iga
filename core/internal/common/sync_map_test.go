package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyncMap(t *testing.T) {
	t.Run("Load misses with the zero value", func(t *testing.T) {
		var m SyncMap[string, *int]

		got, ok := m.Load("absent")

		assert.False(t, ok)
		assert.Nil(t, got)
	})

	t.Run("LoadOrStore stores the first value and returns it on later calls", func(t *testing.T) {
		var m SyncMap[string, *int]
		first, second := new(int), new(int)

		stored, loaded := m.LoadOrStore("k", first)
		assert.False(t, loaded)
		assert.Same(t, first, stored)

		existing, loaded := m.LoadOrStore("k", second)
		assert.True(t, loaded)
		assert.Same(t, first, existing)

		got, ok := m.Load("k")
		assert.True(t, ok)
		assert.Same(t, first, got)
	})
}
